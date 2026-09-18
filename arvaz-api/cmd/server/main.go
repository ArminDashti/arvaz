package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ArminDashti/arvaz-api/internal/asn"
	"github.com/ArminDashti/arvaz-api/internal/auth"
	"github.com/ArminDashti/arvaz-api/internal/config"
	"github.com/ArminDashti/arvaz-api/internal/dockerx"
	"github.com/ArminDashti/arvaz-api/internal/haproxy"
	httpserver "github.com/ArminDashti/arvaz-api/internal/http"
	"github.com/ArminDashti/arvaz-api/internal/metrics"
	"github.com/ArminDashti/arvaz-api/internal/mullvad"
	"github.com/ArminDashti/arvaz-api/internal/softether"
	"github.com/ArminDashti/arvaz-api/internal/store"
	"github.com/ArminDashti/arvaz-api/internal/windscribe"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	var dockerClient *dockerx.Client
	if dc, err := dockerx.New(cfg.PublicIP, cfg.HAProxyConfigPath); err != nil {
		log.Printf("docker client unavailable: %v", err)
	} else {
		dockerClient = dc
		defer dockerClient.Close()
	}

	mvClient := mullvad.New(cfg.MullvadContainer)
	wsClient := windscribe.New(cfg.WindscribeContainer)
	asnResolver := asn.NewAsipResolver(cfg.AsipBaseURL)
	hapClient := haproxy.New("haproxy", "/var/lib/haproxy/admin.sock")
	seClient := softether.New(
		cfg.SoftEtherContainer,
		cfg.SoftEtherPassword,
		cfg.SoftEtherHub,
		cfg.SoftEtherEnabled,
		cfg.SoftEtherVpncmdTimeout,
		asnResolver,
		hapClient,
	)

	ctx := context.Background()
	db, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres required: %v", err)
	}
	defer db.Close()

	authSvc := auth.NewService(db, cfg.JWTSecret)
	if err := authSvc.EnsureDefaultUser(ctx, cfg.DefaultUsername, cfg.DefaultPassword); err != nil {
		log.Fatalf("seed default user: %v", err)
	}

	go pollSoftEther(cfg, seClient, db)
	go pollSoftEtherTraffic(seClient)
	go pollHostMetrics(db)

	srv := httpserver.New(cfg, dockerClient, mvClient, wsClient, seClient, hapClient, authSvc, db)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("arvaz-api listening on http://%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func pollSoftEther(cfg config.Config, se *softether.Client, db *store.Store) {
	if !cfg.SoftEtherEnabled {
		return
	}
	interval := cfg.SoftEtherPollEvery
	if interval < time.Second {
		interval = 120 * time.Second
	}
	var pollMu sync.Mutex
	run := func() {
		if !pollMu.TryLock() {
			log.Printf("softether poll: skipped (previous poll still running)")
			return
		}
		defer pollMu.Unlock()

		// SessionList only (no SessionGet fan-out). Public IPs come from HAProxy
		// via the HTTP handler enrich path on read; poll keeps DB usernames/bytes.
		pollTimeout := cfg.SoftEtherVpncmdTimeout + 30*time.Second
		if pollTimeout < 45*time.Second {
			pollTimeout = 45 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
		defer cancel()
		sessions, err := se.ListSessionTraffic(ctx)
		if err != nil {
			log.Printf("softether poll: %v", err)
			return
		}
		if err := db.SyncOnlineSessions(ctx, sessions); err != nil {
			log.Printf("softether sync: %v", err)
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		run()
	}
}

func pollSoftEtherTraffic(se *softether.Client) {
	if se == nil || !se.Enabled {
		return
	}
	var pollMu sync.Mutex
	run := func() {
		if !pollMu.TryLock() {
			return
		}
		defer pollMu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if _, err := se.ListSessionTraffic(ctx); err != nil {
			log.Printf("softether traffic poll: %v", err)
		}
	}
	run()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		run()
	}
}

func pollHostMetrics(db *store.Store) {
	if db == nil {
		return
	}
	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		snap, err := metrics.Collect()
		if err != nil {
			log.Printf("host metrics poll: %v", err)
			return
		}
		cpu := 0.0
		if len(snap.CPUCores) > 0 {
			sum := 0.0
			for _, v := range snap.CPUCores {
				sum += v
			}
			cpu = sum / float64(len(snap.CPUCores))
		}
		if err := db.InsertHostMetricsSample(ctx, store.HostMetricsSample{
			TS:          snap.Timestamp,
			CPUPct:      cpu,
			MemUsedGB:   snap.Memory.UsedGB,
			MemTotalGB:  snap.Memory.TotalGB,
			DiskUsedGB:  snap.Disk.UsedGB,
			DiskTotalGB: snap.Disk.TotalGB,
			NetDownMbps: snap.Network.DownloadMbps,
			NetUpMbps:   snap.Network.UploadMbps,
		}); err != nil {
			log.Printf("host metrics insert: %v", err)
		}
		if err := db.PruneHostMetricsOlderThan(ctx, time.Now().UTC().Add(-90*24*time.Hour)); err != nil {
			log.Printf("host metrics prune: %v", err)
		}
	}
	run()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		run()
	}
}
