package metrics

import (
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Snapshot is a point-in-time host resource sample for the Docker host OS.
type Snapshot struct {
	Timestamp time.Time     `json:"timestamp"`
	CPUCores  []float64     `json:"cpuCores"`
	Memory    MemoryStats   `json:"memory"`
	Disk      DiskStats     `json:"disk"`
	Network   NetworkStats  `json:"network"`
}

type MemoryStats struct {
	TotalGB     float64 `json:"totalGb"`
	UsedGB      float64 `json:"usedGb"`
	AvailableGB float64 `json:"availableGb"`
}

type DiskStats struct {
	TotalGB float64 `json:"totalGb"`
	UsedGB  float64 `json:"usedGb"`
	FreeGB  float64 `json:"freeGb"`
}

type NetworkStats struct {
	DownloadMbps float64 `json:"downloadMbps"`
	UploadMbps   float64 `json:"uploadMbps"`
}

var (
	netMu   sync.Mutex
	netPrev netCounters
)

// Collect reads host CPU, memory, disk, and network rates.
// CPU uses a non-blocking percent (first call after process start may be zeros).
// Network Mbps uses one in-memory previous counter sample in this process.
func Collect() (Snapshot, error) {
	now := time.Now().UTC()
	out := Snapshot{Timestamp: now}

	cores, err := cpu.Percent(0, true)
	if err != nil {
		return out, err
	}
	out.CPUCores = make([]float64, len(cores))
	for i, v := range cores {
		out.CPUCores[i] = round2(v)
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return out, err
	}
	out.Memory = MemoryStats{
		TotalGB:     bytesToGiB(vm.Total),
		UsedGB:      bytesToGiB(vm.Used),
		AvailableGB: bytesToGiB(vm.Available),
	}

	usage, err := disk.Usage(diskRoot())
	if err != nil {
		return out, err
	}
	out.Disk = DiskStats{
		TotalGB: bytesToGiB(usage.Total),
		UsedGB:  bytesToGiB(usage.Used),
		FreeGB:  bytesToGiB(usage.Free),
	}

	counters, err := readIOCounters()
	if err != nil {
		return out, err
	}
	var recv, sent uint64
	for _, c := range counters {
		if !IncludeNIC(c.Name) {
			continue
		}
		recv += c.BytesRecv
		sent += c.BytesSent
	}

	netMu.Lock()
	dl, ul, ok := MbpsFromDelta(netPrev.bytesRecv, netPrev.bytesSent, netPrev.at, recv, sent, now)
	netPrev = netCounters{bytesRecv: recv, bytesSent: sent, at: now}
	netMu.Unlock()
	if ok {
		out.Network = NetworkStats{
			DownloadMbps: round2(dl),
			UploadMbps:   round2(ul),
		}
	}

	return out, nil
}

func readIOCounters() ([]net.IOCountersStat, error) {
	if path := HostNetDevPath(); path != "" {
		return net.IOCountersByFile(true, path)
	}
	return net.IOCounters(true)
}

func diskRoot() string {
	if r := os.Getenv("HOST_ROOT"); r != "" {
		return r
	}
	if runtime.GOOS == "windows" {
		return `C:\`
	}
	return "/"
}

func bytesToGiB(b uint64) float64 {
	return round2(float64(b) / (1024 * 1024 * 1024))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
