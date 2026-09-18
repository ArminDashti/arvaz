package httpserver

import (
	"net/http"
	"strings"
	"time"

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

func (s *Server) getHostMetricsHistory(c *gin.Context) {
	if s.store == nil {
		c.JSON(http.StatusOK, gin.H{"points": []any{}, "error": "store unavailable"})
		return
	}
	rangeKey := strings.ToLower(strings.TrimSpace(c.Query("range")))
	if rangeKey == "" {
		rangeKey = "today"
	}
	from, to := hostMetricsRange(rangeKey, time.Now())
	points, err := s.store.ListHostMetricsHistory(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"points": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"range": rangeKey, "from": from, "to": to, "points": points})
}

func hostMetricsRange(rangeKey string, now time.Time) (time.Time, time.Time) {
	loc := now.Location()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	switch rangeKey {
	case "yesterday":
		return today.Add(-24 * time.Hour), today
	case "this_week":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := today.AddDate(0, 0, -(weekday - 1))
		return start, now
	case "last_week":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisWeekStart := today.AddDate(0, 0, -(weekday - 1))
		return thisWeekStart.AddDate(0, 0, -7), thisWeekStart
	case "this_month":
		start := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		return start, now
	case "all":
		return time.Unix(0, 0).UTC(), now.Add(time.Second)
	default: // today
		return today, now.Add(time.Second)
	}
}
