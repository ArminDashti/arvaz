package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ArminDashti/arvaz-api/internal/iperf"
	"github.com/gin-gonic/gin"
)

func (s *Server) getIperfLatency(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) getIperfDownload(c *gin.Context) {
	n := iperf.DefaultBytes
	if raw := c.Query("bytes"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bytes"})
			return
		}
		n = parsed
	}
	n = iperf.ClampBytes(n)

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(n, 10))
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusOK)

	_ = iperf.WriteZeros(c.Writer, n)
}

func (s *Server) postIperfUpload(c *gin.Context) {
	limited := http.MaxBytesReader(c.Writer, c.Request.Body, iperf.MaxBytes)
	written, err := iperf.DiscardBody(limited)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bytesReceived": written})
}
