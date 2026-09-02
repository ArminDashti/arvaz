package metrics

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIncludeNIC(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"eth0", true},
		{"ens3", true},
		{"enp3s0", true},
		{"wlan0", true},
		{"lo", false},
		{"LO", false},
		{"veth1234ab", false},
		{"docker0", false},
		{"br-abcdef", false},
		{"amn0", false},
		{"tap0", false},
		{"tun0", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IncludeNIC(tc.name); got != tc.want {
			t.Errorf("IncludeNIC(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHostNetDevPath(t *testing.T) {
	t.Setenv("HOST_PROC", "")
	if got := HostNetDevPath(); got != "" {
		t.Fatalf("empty HOST_PROC: got %q", got)
	}
	t.Setenv("HOST_PROC", "/host/proc")
	got := HostNetDevPath()
	want := filepath.Join("/host/proc", "1", "net", "dev")
	if got != want {
		t.Fatalf("HostNetDevPath() = %q, want %q", got, want)
	}
}

func TestMbpsFromDelta(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 27, 12, 0, 2, 0, time.UTC)
	prevAt := now.Add(-2 * time.Second)

	// 2 seconds, +2_500_000 bytes recv => 10 Mbps; +1_250_000 sent => 5 Mbps
	dl, ul, ok := MbpsFromDelta(0, 0, prevAt, 2_500_000, 1_250_000, now)
	if !ok {
		t.Fatal("expected ok")
	}
	if dl < 9.9 || dl > 10.1 {
		t.Errorf("downloadMbps = %v, want ~10", dl)
	}
	if ul < 4.9 || ul > 5.1 {
		t.Errorf("uploadMbps = %v, want ~5", ul)
	}

	_, _, ok = MbpsFromDelta(0, 0, time.Time{}, 100, 100, now)
	if ok {
		t.Error("missing prev should not be ok")
	}

	_, _, ok = MbpsFromDelta(0, 0, now.Add(-40*time.Second), 100, 100, now)
	if ok {
		t.Error("stale prev should not be ok")
	}

	_, _, ok = MbpsFromDelta(200, 200, prevAt, 100, 100, now)
	if ok {
		t.Error("counter wrap should not be ok")
	}
}
