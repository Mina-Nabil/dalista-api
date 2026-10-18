package middlewares

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type rateVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

const (
	rateVisitorIdleTTL    = 10 * time.Minute
	rateVisitorSweepEvery = 5 * time.Minute
	rateRetryAfterSeconds = 60
)

// RateLimitMiddleware throttles each client IP to perMinute requests per minute
// using a token bucket (burst = perMinute, refill = perMinute/60 per second).
// Returns a no-op handler when perMinute <= 0.
//
// Buckets are kept in a process-local sync.Map; a janitor goroutine evicts idle
// visitors so memory does not grow unbounded.
func RateLimitMiddleware(perMinute int) gin.HandlerFunc {
	if perMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	var visitors sync.Map
	limit := rate.Limit(float64(perMinute) / 60.0)
	burst := perMinute

	go func() {
		ticker := time.NewTicker(rateVisitorSweepEvery)
		defer ticker.Stop()
		for now := range ticker.C {
			visitors.Range(func(key, value any) bool {
				if v, ok := value.(*rateVisitor); ok && now.Sub(v.lastSeen) > rateVisitorIdleTTL {
					visitors.Delete(key)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		var v *rateVisitor
		if existing, ok := visitors.Load(ip); ok {
			v = existing.(*rateVisitor)
		} else {
			fresh := &rateVisitor{limiter: rate.NewLimiter(limit, burst), lastSeen: now}
			actual, _ := visitors.LoadOrStore(ip, fresh)
			v = actual.(*rateVisitor)
		}
		v.lastSeen = now

		if !v.limiter.Allow() {
			c.Header("Retry-After", strconv.Itoa(rateRetryAfterSeconds))
			c.JSON(http.StatusTooManyRequests, gin.H{"message": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
