package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods  = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders  = "Authorization, Content-Type, X-Requested-With, X-Cron-Secret"
	corsExposeHeaders = "Content-Disposition, Retry-After"
	corsMaxAge        = "86400"
)

type OriginMatcher struct {
	exact    map[string]struct{}
	patterns []*regexp.Regexp
}

func NewOriginMatcher(origins []string, patterns []*regexp.Regexp) *OriginMatcher {
	exact := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		exact[o] = struct{}{}
	}
	return &OriginMatcher{exact: exact, patterns: patterns}
}

func (m *OriginMatcher) Allowed(origin string) bool {
	if _, ok := m.exact[origin]; ok {
		return true
	}
	for _, re := range m.patterns {
		if re.MatchString(origin) {
			return true
		}
	}
	return false
}

// CORS wajib dipasang paling luar: header harus ikut di response error dan panic,
// dan preflight dijawab sebelum auth atau rate limit berjalan.
func CORS(m *OriginMatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Add("Vary", "Origin")

		origin := c.GetHeader("Origin")
		allowed := origin != "" && m.Allowed(origin)
		if allowed {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Expose-Headers", corsExposeHeaders)
		}

		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			if allowed {
				h.Set("Access-Control-Allow-Methods", corsAllowMethods)
				h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
				h.Set("Access-Control-Max-Age", corsMaxAge)
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
