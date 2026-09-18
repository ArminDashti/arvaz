package httpserver

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ArminDashti/arvaz-api/internal/asn"
	"github.com/ArminDashti/arvaz-api/internal/auth"
	"github.com/ArminDashti/arvaz-api/internal/config"
	"github.com/ArminDashti/arvaz-api/internal/dockerx"
	"github.com/ArminDashti/arvaz-api/internal/haproxy"
	"github.com/ArminDashti/arvaz-api/internal/mullvad"
	"github.com/ArminDashti/arvaz-api/internal/softether"
	"github.com/ArminDashti/arvaz-api/internal/windscribe"
	"github.com/ArminDashti/arvaz-api/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg       config.Config
	docker    *dockerx.Client
	mullvad    *mullvad.Client
	windscribe *windscribe.Client
	softether  *softether.Client
	haproxy   *haproxy.Client
	auth      *auth.Service
	store     *store.Store
}

func New(cfg config.Config, d *dockerx.Client, mv *mullvad.Client, ws *windscribe.Client, se *softether.Client, hap *haproxy.Client, authSvc *auth.Service, st *store.Store) *Server {
	return &Server{cfg: cfg, docker: d, mullvad: mv, windscribe: ws, softether: se, haproxy: hap, auth: authSvc, store: st}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/auth/login", s.postLogin)
	}

	protected := v1.Group("")
	protected.Use(jwtMiddleware(s.auth))
	{
		protected.GET("/docker/containers", s.getDockerContainers)
		protected.POST("/docker/containers/:name/start", s.postDockerStart)
		protected.POST("/docker/containers/:name/stop", s.postDockerStop)
		protected.POST("/docker/containers/:name/exec", s.postDockerExec)
		protected.GET("/mullvad/status", s.getMullvadStatus)
		protected.GET("/mullvad/relays", s.getMullvadRelays)
		protected.POST("/mullvad/relay", s.postMullvadRelay)
		protected.GET("/mullvad/anti-censorship", s.getMullvadAnti)
		protected.POST("/mullvad/anti-censorship", s.postMullvadAnti)
		protected.POST("/mullvad/tunnel", s.postMullvadTunnel)
		protected.POST("/mullvad/ping", s.postMullvadPing)
		protected.POST("/mullvad/speedtest", s.postMullvadSpeedtest)
		protected.GET("/windscribe/status", s.getWindscribeStatus)
		protected.GET("/windscribe/locations", s.getWindscribeLocations)
		protected.POST("/windscribe/connect", s.postWindscribeConnect)
		protected.POST("/windscribe/disconnect", s.postWindscribeDisconnect)
		protected.POST("/windscribe/ping", s.postWindscribePing)
		protected.POST("/windscribe/speedtest", s.postWindscribeSpeedtest)
		protected.GET("/system/host-metrics", s.getHostMetrics)
		protected.GET("/system/host-metrics/history", s.getHostMetricsHistory)
		protected.GET("/iperf/latency", s.getIperfLatency)
		protected.GET("/iperf/download", s.getIperfDownload)
		protected.POST("/iperf/upload", s.postIperfUpload)
		protected.GET("/softether/sessions", s.getSoftEtherSessions)
		protected.GET("/softether/ip-sessions", s.getSoftEtherIpSessions)
		protected.GET("/softether/users", s.getSoftEtherUsers)
		protected.GET("/softether/users/:username/sessions", s.getSoftEtherUserSessions)
	}
	return r
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) postLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := s.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) getDockerContainers(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"containers": []any{}, "error": "docker unavailable"})
		return
	}
	containers, err := s.docker.ListContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"containers": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"containers": containers})
}

func (s *Server) postDockerStart(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "docker unavailable"})
		return
	}
	name := c.Param("name")
	if err := s.docker.StartContainer(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postDockerStop(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "docker unavailable"})
		return
	}
	name := c.Param("name")
	if err := s.docker.StopContainer(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postDockerExec(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "docker unavailable"})
		return
	}
	var req struct {
		Command string `json:"command"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Command) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "command required"})
		return
	}
	out, err := s.docker.ExecShell(c.Request.Context(), c.Param("name"), req.Command)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "output": out, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "output": out})
}

func (s *Server) getMullvadStatus(c *gin.Context) {
	st, err := s.mullvad.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": st})
}

func (s *Server) getMullvadRelays(c *gin.Context) {
	relays, err := s.mullvad.ListRelays(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"relays": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"relays": relays})
}

func (s *Server) postMullvadRelay(c *gin.Context) {
	var req struct {
		Country  string `json:"country" binding:"required"`
		City     string `json:"city"`
		Hostname string `json:"hostname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := s.mullvad.SetRelay(c.Request.Context(), req.Country, req.City, req.Hostname); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) getMullvadAnti(c *gin.Context) {
	mode, err := s.mullvad.GetAntiCensorship(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mode": mode})
}

func (s *Server) postMullvadAnti(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := s.mullvad.SetAntiCensorship(c.Request.Context(), req.Mode); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postMullvadTunnel(c *gin.Context) {
	var req struct {
		QuantumResistant *bool `json:"quantumResistant"`
		Daita            *bool `json:"daita"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.QuantumResistant == nil && req.Daita == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantumResistant or daita required"})
		return
	}
	if err := s.mullvad.SetTunnel(c.Request.Context(), req.QuantumResistant, req.Daita); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postMullvadPing(c *gin.Context) {
	var req struct {
		Target string `json:"target"`
		Count  int    `json:"count"`
	}
	_ = c.ShouldBindJSON(&req)
	res, err := s.mullvad.Ping(c.Request.Context(), req.Target, req.Count)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *Server) postMullvadSpeedtest(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	_ = c.ShouldBindJSON(&req)
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "parallel"
	}
	res, err := s.mullvad.Speedtest(c.Request.Context(), mode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *Server) getWindscribeStatus(c *gin.Context) {
	st, err := s.windscribe.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": st})
}

func (s *Server) getWindscribeLocations(c *gin.Context) {
	locs, err := s.windscribe.ListLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"locations": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locs})
}

func (s *Server) postWindscribeConnect(c *gin.Context) {
	var req struct {
		Location string `json:"location" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := s.windscribe.Connect(c.Request.Context(), req.Location); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postWindscribeDisconnect(c *gin.Context) {
	if err := s.windscribe.Disconnect(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postWindscribePing(c *gin.Context) {
	var req struct {
		Target string `json:"target"`
		Count  int    `json:"count"`
	}
	_ = c.ShouldBindJSON(&req)
	res, err := s.windscribe.Ping(c.Request.Context(), req.Target, req.Count)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *Server) postWindscribeSpeedtest(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	_ = c.ShouldBindJSON(&req)
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "parallel"
	}
	res, err := s.windscribe.Speedtest(c.Request.Context(), mode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *Server) getSoftEtherSessions(c *gin.Context) {
	ctx := c.Request.Context()
	// Serve from DB cache (background poll). Live vpncmd on every page load
	// stacks concurrent SessionList orphans and pegs CPU on T3.
	var liveErr error
	if s.store != nil {
		sessions, err := s.store.ListOnlineSessions(ctx)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
			return
		}
		if len(sessions) == 0 || softEtherAllZeroTraffic(sessions) {
			live, err := s.tryLiveSoftEtherSessions(ctx)
			if err != nil {
				liveErr = err
			}
			if len(live) > 0 {
				sessions = live
			}
		}
		if len(sessions) == 0 {
			sessions = s.softEtherSessionsFromHAProxy(ctx)
		}
		s.enrichSoftEtherPublicIPs(ctx, sessions)
		s.enrichSoftEtherSessions(ctx, sessions)
		resp := gin.H{"sessions": sessions}
		if liveErr != nil && len(sessions) == 0 {
			resp["error"] = liveErr.Error()
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	if s.softether == nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": "softether unavailable"})
		return
	}
	sessions, err := s.softether.ListOnlineSessions(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	s.enrichSoftEtherSessions(ctx, sessions)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func softEtherAllZeroTraffic(sessions []softether.OnlineSession) bool {
	if len(sessions) == 0 {
		return true
	}
	for _, s := range sessions {
		if s.DownloadBytes > 0 || s.UploadBytes > 0 {
			return false
		}
	}
	return true
}

func (s *Server) enrichSoftEtherSessions(ctx context.Context, sessions []softether.OnlineSession) {
	for i := range sessions {
		row := &sessions[i]
		if (row.Username == "" || row.Username == row.ClientIP) && row.ClientIP != "" && s.store != nil {
			if hist, err := s.store.LookupLatestByClientIP(ctx, row.ClientIP); err == nil && hist != nil {
				if hist.Username != "" && hist.Username != row.ClientIP {
					row.Username = hist.Username
				}
			}
		}
		if row.Username == row.ClientIP {
			row.Username = ""
		}
		row.LastISP, row.IspLogo = asn.WithLogo(row.LastISP)
	}
	if s.softether != nil {
		s.softether.AttachRates(sessions)
	}
}

func (s *Server) tryLiveSoftEtherSessions(ctx context.Context) ([]softether.OnlineSession, error) {
	if s.softether == nil || !s.softether.Enabled {
		return nil, nil
	}
	// SessionList only — never SessionGet here. Per-session vpncmd under the
	// global lock stacked 1s UI polls into multi-hour hangs on T3.
	liveCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	sessions, err := s.softether.ListSessionTraffic(liveCtx)
	if err != nil {
		return nil, err
	}
	s.enrichSoftEtherPublicIPs(liveCtx, sessions)
	if len(sessions) > 0 && s.store != nil {
		if syncErr := s.store.SyncOnlineSessions(liveCtx, sessions); syncErr != nil {
			log.Printf("softether live sync: %v", syncErr)
		}
	}
	return sessions, nil
}

// enrichSoftEtherPublicIPs fills missing public client IPs from softether_user_stats
// (last known IP per username). Avoids HAProxy show-sess fan-out on the 1s UI poll path.
func (s *Server) enrichSoftEtherPublicIPs(ctx context.Context, sessions []softether.OnlineSession) {
	if s.store == nil || len(sessions) == 0 {
		return
	}
	stats, err := s.store.GetUserStatMap(ctx)
	if err != nil || len(stats) == 0 {
		return
	}
	for i := range sessions {
		row := &sessions[i]
		if row.ClientIP != "" || row.Username == "" {
			continue
		}
		st, ok := stats[row.Username]
		if !ok || st.ClientIP == "" {
			continue
		}
		row.ClientIP = st.ClientIP
		if row.LastISP == "" && st.ISP != "" {
			row.LastISP = st.ISP
		}
	}
	if s.softether == nil {
		return
	}
	for i := range sessions {
		row := &sessions[i]
		if row.ClientIP == "" || row.LastISP != "" {
			continue
		}
		if label := s.softether.ASN.Lookup(row.ClientIP); label != "" {
			row.LastISP = label
		}
	}
}

func (s *Server) softEtherSessionsFromHAProxy(ctx context.Context) []softether.OnlineSession {
	if s.haproxy == nil || s.store == nil {
		return nil
	}
	entries, err := s.haproxy.ListFrontendSessions(ctx)
	if err != nil || len(entries) == 0 {
		return nil
	}

	// Best-effort live username/traffic map keyed by public client IP.
	liveByIP := map[string]softether.OnlineSession{}
	liveByUser := map[string]softether.OnlineSession{}
	if s.softether != nil && s.softether.Enabled {
		liveCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		live, liveErr := s.softether.ListSessionTraffic(liveCtx)
		cancel()
		if liveErr == nil {
			s.enrichSoftEtherPublicIPs(ctx, live)
			for _, row := range live {
				if row.ClientIP != "" {
					liveByIP[row.ClientIP] = row
				}
				if row.Username != "" {
					liveByUser[strings.ToLower(row.Username)] = row
				}
			}
		}
	}

	now := time.Now().UTC()
	out := make([]softether.OnlineSession, 0, len(entries))
	seen := map[string]struct{}{}
	for _, e := range entries {
		if e.IP == "" {
			continue
		}
		if _, ok := seen[e.IP]; ok {
			continue
		}
		seen[e.IP] = struct{}{}
		row := softether.OnlineSession{ClientIP: e.IP}
		if live, ok := liveByIP[e.IP]; ok {
			row = live
			if row.ClientIP == "" {
				row.ClientIP = e.IP
			}
		} else if hist, err := s.store.LookupLatestByClientIP(ctx, e.IP); err == nil && hist != nil {
			row.Username = hist.Username
			row.DownloadBytes = hist.DownloadBytes
			row.UploadBytes = hist.UploadBytes
			if hist.ISP != "" {
				row.LastISP = hist.ISP
			}
			if live, ok := liveByUser[strings.ToLower(hist.Username)]; ok {
				row.DownloadBytes = live.DownloadBytes
				row.UploadBytes = live.UploadBytes
				row.TransferBytes = live.TransferBytes
				row.SessionName = live.SessionName
				row.SessionKey = live.SessionKey
				if live.DownloadMbps != nil {
					row.DownloadMbps = live.DownloadMbps
				}
				if live.UploadMbps != nil {
					row.UploadMbps = live.UploadMbps
				}
			}
		}
		if row.Username == "" || row.Username == e.IP {
			// Do not surface raw IP as username when we cannot resolve a hub account.
			continue
		}
		if row.ConnectedAt == nil {
			connected := now.Add(-e.Age)
			row.ConnectedAt = &connected
			row.SessionDurationSeconds = int64(e.Age.Seconds())
		}
		if row.SessionKey == "" {
			row.SessionKey = row.Username + "|haproxy|" + e.IP
		}
		out = append(out, row)
	}
	return out
}

func (s *Server) getSoftEtherUsers(c *gin.Context) {
	ctx := c.Request.Context()
	var users []softether.HubUser
	var liveErr error

	if s.softether != nil {
		users, liveErr = s.softether.ListUsers(ctx)
	}
	if len(users) == 0 && s.store != nil {
		if cached, err := s.store.ListHubUsersFromStats(ctx); err == nil && len(cached) > 0 {
			users = cached
			liveErr = nil
		}
	}
	if len(users) == 0 && s.softether == nil {
		c.JSON(http.StatusOK, gin.H{"users": []any{}, "error": "softether unavailable"})
		return
	}

	stats := map[string]store.UserStat{}
	periods := map[string]store.UserTrafficPeriods{}
	if s.store != nil {
		if m, err := s.store.GetUserStatMap(ctx); err == nil {
			stats = m
		}
		if m, err := s.store.GetUserTrafficPeriodsMap(ctx); err == nil {
			periods = m
		}
	}

	out := make([]softether.HubUser, 0, len(users))
	for _, u := range users {
		row := u
		row.LastISP, row.IspLogo = asn.WithLogo(row.LastISP)
		if st, ok := stats[u.Username]; ok {
			if st.DownloadBytes > row.DownloadBytes {
				row.DownloadBytes = st.DownloadBytes
			}
			if st.UploadBytes > row.UploadBytes {
				row.UploadBytes = st.UploadBytes
			}
			if row.LastIP == "" {
				row.LastIP = st.ClientIP
			}
			if row.LastISP == "" {
				row.LastISP, row.IspLogo = asn.WithLogo(st.ISP)
			}
		}
		// Prefer live public IP from online enrichment when stats still hold a private bridge IP.
		if row.LastIP != "" && strings.HasPrefix(row.LastIP, "172.") {
			row.LastIP = ""
		}
		if p, ok := periods[u.Username]; ok {
			row.TrafficYesterdayBytes = p.YesterdayBytes
			row.TrafficWeekBytes = p.WeekBytes
			row.TrafficMonthBytes = p.MonthBytes
			row.TrafficYesterdaySeconds = p.YesterdaySeconds
			row.TrafficWeekSeconds = p.WeekSeconds
			row.TrafficMonthSeconds = p.MonthSeconds
		}
		row.TrafficTotalBytes = row.DownloadBytes + row.UploadBytes
		out = append(out, row)
	}
	resp := gin.H{"users": out}
	if liveErr != nil && len(out) == 0 {
		resp["error"] = liveErr.Error()
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) getSoftEtherUserSessions(c *gin.Context) {
	if s.store == nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": "store unavailable"})
		return
	}
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}
	sessions, err := s.store.ListSessionsByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	for i := range sessions {
		sessions[i].ISP, sessions[i].IspLogo = asn.WithLogo(sessions[i].ISP)
	}
	c.JSON(http.StatusOK, gin.H{"username": username, "sessions": sessions})
}

func (s *Server) getSoftEtherIpSessions(c *gin.Context) {
	if s.store == nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": "store unavailable"})
		return
	}
	ip := strings.TrimSpace(c.Query("ip"))
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip required"})
		return
	}
	sessions, err := s.store.ListSessionsByIP(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ip": ip, "sessions": []any{}, "error": err.Error()})
		return
	}
	for i := range sessions {
		sessions[i].ISP, sessions[i].IspLogo = asn.WithLogo(sessions[i].ISP)
	}
	c.JSON(http.StatusOK, gin.H{"ip": ip, "sessions": sessions})
}
