package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name   string
		header map[string]string
		want   string
	}{
		{"xff berantai", map[string]string{"X-Forwarded-For": "203.0.113.7, 10.0.0.1"}, "203.0.113.7"},
		{"xff tunggal", map[string]string{"X-Forwarded-For": "  203.0.113.8  "}, "203.0.113.8"},
		{"x-real-ip", map[string]string{"X-Real-IP": " 203.0.113.9 "}, "203.0.113.9"},
		{"xff menang atas x-real-ip", map[string]string{"X-Forwarded-For": "203.0.113.7", "X-Real-IP": "203.0.113.9"}, "203.0.113.7"},
		{"tanpa header", nil, "192.0.2.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Request.RemoteAddr = "192.0.2.1:4321"
			for k, v := range tc.header {
				c.Request.Header.Set(k, v)
			}
			if got := ClientIP(c); got != tc.want {
				t.Errorf("ClientIP = %q, want %q", got, tc.want)
			}
		})
	}
}
