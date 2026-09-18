package windscribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cliBin = "windscribe-cli"

type Client struct {
	Container string
}

func New(container string) *Client {
	if strings.TrimSpace(container) == "" {
		container = "windscribe"
	}
	return &Client{Container: container}
}

type Status struct {
	Raw       string `json:"raw"`
	Connected bool   `json:"connected"`
	Location  string `json:"location,omitempty"`
	PublicIP  string `json:"publicIp,omitempty"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
}

type Location struct {
	Region      string `json:"region"`
	City        string `json:"city"`
	Nickname    string `json:"nickname"`
	Label       string `json:"label"`
	CountryCode string `json:"countryCode,omitempty"`
	Active      bool   `json:"active"`
}

type PingResult struct {
	Target            string  `json:"target"`
	Count             int     `json:"count"`
	PacketLossPercent float64 `json:"packetLossPercent"`
	AvgMs             float64 `json:"avgMs"`
	Raw               string  `json:"raw,omitempty"`
}

type SpeedtestResult struct {
	Mode         string  `json:"mode"`
	Raw          string  `json:"raw"`
	DownloadMbps float64 `json:"downloadMbps,omitempty"`
	UploadMbps   float64 `json:"uploadMbps,omitempty"`
	LatencyMs    float64 `json:"latencyMs,omitempty"`
	ParsedOK     bool    `json:"parsedOk"`
}

var (
	reConnectState = regexp.MustCompile(`(?i)Connect state:\s*(?:\*)?(Connected|Disconnected)(?::\s*(.+))?`)
	reProtocol     = regexp.MustCompile(`(?i)Protocol:\s*(.+)`)
	reStatusIP     = regexp.MustCompile(`(?i)(?:External\s+)?IP(?:\s+address)?\s*:\s*([0-9.]+)`)
	reISORegion    = regexp.MustCompile(`^[A-Za-z]{2}$`)
)

func (c *Client) Status(ctx context.Context) (*Status, error) {
	raw, err := c.exec(ctx, 20*time.Second, cliBin, "status")
	if err != nil {
		return nil, err
	}
	st := parseStatus(raw)
	return st, nil
}

func parseStatus(raw string) *Status {
	st := &Status{Raw: strings.TrimSpace(raw)}
	if m := reConnectState.FindStringSubmatch(raw); len(m) >= 2 {
		st.Connected = strings.EqualFold(strings.TrimSpace(m[1]), "Connected")
		if len(m) >= 3 {
			st.Location = strings.TrimSpace(m[2])
		}
	}
	if m := reProtocol.FindStringSubmatch(raw); len(m) == 2 {
		st.Protocol = strings.TrimSpace(m[1])
	}
	if m := reStatusIP.FindStringSubmatch(raw); len(m) == 2 {
		st.PublicIP = strings.TrimSpace(m[1])
	}
	if st.Location != "" {
		parts := splitLocationLabel(st.Location)
		if len(parts) >= 1 {
			st.Country = parts[0]
		}
		if len(parts) >= 2 {
			st.City = parts[1]
		}
	}
	return st
}

func splitLocationLabel(label string) []string {
	parts := strings.Split(label, " - ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (c *Client) ListLocations(ctx context.Context) ([]Location, error) {
	raw, err := c.exec(ctx, 60*time.Second, cliBin, "locations")
	if err != nil {
		return nil, err
	}
	status, _ := c.Status(ctx)
	activeLabel := ""
	if status != nil {
		activeLabel = status.Location
	}
	return parseLocationList(raw, activeLabel), nil
}

func parseLocationList(raw, activeLabel string) []Location {
	activeLabel = strings.TrimSpace(strings.TrimPrefix(activeLabel, "*"))
	out := make([]Location, 0, 128)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		active := strings.HasPrefix(line, "*")
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		parts := splitLocationLabel(line)
		if len(parts) < 2 {
			continue
		}
		loc := Location{
			Region:   parts[0],
			Label:    line,
			Active:   active || locationMatchesActive(line, parts, activeLabel),
			Nickname: parts[len(parts)-1],
		}
		if len(parts) >= 2 {
			loc.City = parts[1]
		}
		if reISORegion.MatchString(loc.Region) {
			loc.CountryCode = strings.ToLower(loc.Region)
		}
		out = append(out, loc)
	}
	return out
}

func locationMatchesActive(label string, parts []string, active string) bool {
	if active == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(label), active) {
		return true
	}
	if len(parts) >= 2 && strings.EqualFold(parts[0]+" - "+parts[1], active) {
		return true
	}
	return strings.Contains(strings.ToLower(active), strings.ToLower(parts[0]))
}

func (c *Client) Connect(ctx context.Context, location string) error {
	location = strings.TrimSpace(location)
	if location == "" {
		return fmt.Errorf("location required")
	}
	_, err := c.exec(ctx, 90*time.Second, cliBin, "connect", location)
	return err
}

func (c *Client) Disconnect(ctx context.Context) error {
	_, err := c.exec(ctx, 45*time.Second, cliBin, "disconnect")
	return err
}

func (c *Client) Ping(ctx context.Context, target string, count int) (*PingResult, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		target = "1.1.1.1"
	}
	if !validPingTarget(target) {
		return nil, fmt.Errorf("invalid ping target")
	}
	if count < 1 {
		count = 4
	}
	if count > 128 {
		count = 128
	}
	timeout := time.Duration(count*2+10) * time.Second
	if timeout < 20*time.Second {
		timeout = 20 * time.Second
	}
	raw, err := c.exec(ctx, timeout, "ping", "-c", strconv.Itoa(count), target)
	if err != nil && raw == "" {
		return nil, err
	}
	loss, avg := parsePingStats(raw)
	return &PingResult{
		Target:            target,
		Count:             count,
		PacketLossPercent: loss,
		AvgMs:             avg,
		Raw:               strings.TrimSpace(raw),
	}, nil
}

var rePingTarget = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)
var rePingLoss = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)%\s*packet\s+loss`)
var rePingAvg = regexp.MustCompile(`(?i)(?:rtt|round-trip)[^=]*=\s*[0-9.]+/([0-9.]+)`)

func validPingTarget(s string) bool {
	return rePingTarget.MatchString(s)
}

func parsePingStats(raw string) (lossPercent, avgMs float64) {
	if m := rePingLoss.FindStringSubmatch(raw); len(m) == 2 {
		lossPercent, _ = strconv.ParseFloat(m[1], 64)
	}
	if m := rePingAvg.FindStringSubmatch(raw); len(m) == 2 {
		avgMs, _ = strconv.ParseFloat(m[1], 64)
	}
	return lossPercent, avgMs
}

func (c *Client) Speedtest(ctx context.Context, mode string) (*SpeedtestResult, error) {
	mode = strings.TrimSpace(strings.ToLower(mode))
	args := []string{"speedtest", "--accept-license", "--accept-gdpr", "--progress=no", "-f", "json"}
	if mode == "single" {
		args = append(args, "--single")
	} else {
		mode = "parallel"
	}
	raw, err := c.exec(ctx, 120*time.Second, args...)
	if err != nil && raw == "" {
		return nil, err
	}
	res := &SpeedtestResult{Mode: mode, Raw: strings.TrimSpace(raw)}
	parseSpeedtestJSON(res)
	if !res.ParsedOK {
		parseSpeedtestSimple(res)
	}
	return res, nil
}

type ooklaSpeedtestJSON struct {
	Ping     json.RawMessage `json:"ping"`
	Download json.RawMessage `json:"download"`
	Upload   json.RawMessage `json:"upload"`
}

type ooklaBandwidth struct {
	Bandwidth float64 `json:"bandwidth"`
}

type ooklaPing struct {
	Latency float64 `json:"latency"`
}

func parseSpeedtestJSON(res *SpeedtestResult) {
	raw := strings.TrimSpace(res.Raw)
	if raw == "" {
		return
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}
	blob := raw[start : end+1]
	var parsed ooklaSpeedtestJSON
	if err := json.Unmarshal([]byte(blob), &parsed); err != nil {
		return
	}
	dlBw, dlOk := parseOoklaBandwidth(parsed.Download)
	ulBw, ulOk := parseOoklaBandwidth(parsed.Upload)
	lat, latOk := parseOoklaLatency(parsed.Ping)
	if dlOk || ulOk || latOk {
		if dlOk {
			res.DownloadMbps = dlBw
		}
		if ulOk {
			res.UploadMbps = ulBw
		}
		if latOk {
			res.LatencyMs = lat
		}
		res.ParsedOK = true
	}
}

func parseOoklaBandwidth(raw json.RawMessage) (mbps float64, ok bool) {
	raw = json.RawMessage(bytes.TrimSpace(raw))
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	if raw[0] == '{' {
		var stream ooklaBandwidth
		if err := json.Unmarshal(raw, &stream); err != nil {
			return 0, false
		}
		if stream.Bandwidth <= 0 {
			return 0, false
		}
		return stream.Bandwidth * 8 / 1_000_000, true
	}
	var bits float64
	if err := json.Unmarshal(raw, &bits); err != nil || bits <= 0 {
		return 0, false
	}
	return bits / 1_000_000, true
}

func parseOoklaLatency(raw json.RawMessage) (ms float64, ok bool) {
	raw = json.RawMessage(bytes.TrimSpace(raw))
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	if raw[0] == '{' {
		var ping ooklaPing
		if err := json.Unmarshal(raw, &ping); err != nil {
			return 0, false
		}
		if ping.Latency <= 0 {
			return 0, false
		}
		return ping.Latency, true
	}
	var ping float64
	if err := json.Unmarshal(raw, &ping); err != nil || ping <= 0 {
		return 0, false
	}
	return ping, true
}

func parseSpeedtestSimple(res *SpeedtestResult) {
	re := regexp.MustCompile(`(?i)(ping|latency|download|upload):\s*([0-9.]+)`)
	for _, line := range strings.Split(res.Raw, "\n") {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) < 3 {
			continue
		}
		v, _ := strconv.ParseFloat(m[2], 64)
		switch strings.ToLower(m[1]) {
		case "ping", "latency":
			res.LatencyMs = v
		case "download":
			res.DownloadMbps = v
		case "upload":
			res.UploadMbps = v
		}
	}
	res.ParsedOK = res.DownloadMbps > 0 || res.UploadMbps > 0 || res.LatencyMs > 0
}

func (c *Client) exec(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	for _, a := range args {
		lower := strings.ToLower(a)
		if strings.Contains(lower, "logout") {
			return "", fmt.Errorf("forbidden windscribe operation")
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmdArgs := append([]string{"exec", c.Container}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(out)
		}
		if msg == "" {
			msg = err.Error()
		}
		if out != "" {
			return out, fmt.Errorf("%s", msg)
		}
		return "", fmt.Errorf("%s", msg)
	}
	return out, nil
}
