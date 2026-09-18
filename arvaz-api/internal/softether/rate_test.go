package softether

import (
	"testing"
	"time"
)

func TestLooksLikeIPv4(t *testing.T) {
	if !looksLikeIPv4("1.2.3.4") {
		t.Fatal("expected IPv4")
	}
	if looksLikeIPv4("alice") {
		t.Fatal("username should not look like IP")
	}
}

func TestSanitizeSessionIdentity(t *testing.T) {
	s := OnlineSession{Username: "5.6.7.8", ClientIP: "5.6.7.8"}
	sanitizeSessionIdentity(&s)
	if s.Username != "" {
		t.Fatalf("username=%q want empty", s.Username)
	}
}

func TestRateTrackerMbps(t *testing.T) {
	var r rateTracker
	a := []OnlineSession{{SessionName: "sid1", DownloadBytes: 1_000_000, UploadBytes: 500_000}}
	r.apply(a)
	r.mu.Lock()
	sample := r.prev["sid1"]
	sample.at = sample.at.Add(-time.Second)
	r.prev["sid1"] = sample
	r.mu.Unlock()

	b := []OnlineSession{{SessionName: "sid1", DownloadBytes: 2_000_000, UploadBytes: 1_000_000}}
	r.apply(b)
	if b[0].DownloadMbps == nil || *b[0].DownloadMbps <= 0 {
		t.Fatalf("downloadMbps=%v", b[0].DownloadMbps)
	}
	if b[0].UploadMbps == nil || *b[0].UploadMbps <= 0 {
		t.Fatalf("uploadMbps=%v", b[0].UploadMbps)
	}
}
