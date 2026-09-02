package httpserver

import (
	"net/http"

	"github.com/ArminDashti/arvaz-api/internal/metrics"
	"github.com/gin-gonic/gin"
)

func (s *Server) getHostMetrics(c *gin.Context) {
	snap, err := metrics.Collect()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}
