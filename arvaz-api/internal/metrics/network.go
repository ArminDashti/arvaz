package metrics

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const staleNetworkSample = 30 * time.Second

type netCounters struct {
	bytesRecv uint64
	bytesSent uint64
	at        time.Time
}

// IncludeNIC reports whether an interface should contribute to host network rates.
// Skips loopback and common virtual/docker/SoftEther adapters (sum of remaining NICs).
func IncludeNIC(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" || n == "lo" {
		return false
	}
	if strings.HasPrefix(n, "lo.") || strings.HasPrefix(n, "lo:") {
		return false
	}
	if strings.HasPrefix(n, "veth") {
		return false
	}
	if strings.HasPrefix(n, "docker") {
		return false
	}
	if strings.HasPrefix(n, "br-") {
		return false
	}
	// SoftEther / VPN virtual NICs (e.g. amn0) and tunnel/tap devices.
	if strings.HasPrefix(n, "amn") || strings.HasPrefix(n, "tap") || strings.HasPrefix(n, "tun") {
		return false
	}
	if strings.HasPrefix(n, "virbr") || strings.HasPrefix(n, "vnet") {
		return false
	}
	return true
}

// HostNetDevPath returns the /proc/.../net/dev path for host-namespace counters.
// When HOST_PROC is set (API in Docker), /proc/net follows the container netns, so
// we read host init's netns via HOST_PROC/1/net/dev instead.
func HostNetDevPath() string {
	hostProc := strings.TrimSpace(os.Getenv("HOST_PROC"))
	if hostProc == "" {
		return ""
	}
	return filepath.Join(hostProc, "1", "net", "dev")
}

// MbpsFromDelta converts byte counters over elapsed time to megabits per second.
// Returns 0 when previous is missing, elapsed is non-positive, sample is stale, or counters wrapped.
func MbpsFromDelta(prevRecv, prevSent uint64, prevAt time.Time, recv, sent uint64, now time.Time) (downloadMbps, uploadMbps float64, ok bool) {
	if prevAt.IsZero() {
		return 0, 0, false
	}
	elapsed := now.Sub(prevAt)
	if elapsed <= 0 || elapsed > staleNetworkSample {
		return 0, 0, false
	}
	if recv < prevRecv || sent < prevSent {
		return 0, 0, false
	}
	sec := elapsed.Seconds()
	downloadMbps = float64(recv-prevRecv) / sec * 8 / 1e6
	uploadMbps = float64(sent-prevSent) / sec * 8 / 1e6
	return downloadMbps, uploadMbps, true
}
