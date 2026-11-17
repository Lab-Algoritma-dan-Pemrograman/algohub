package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCronAuthorized(t *testing.T) {
	cases := []struct {
		name   string
		secret string
		header map[string]string
		want   bool
	}{
		{"bearer cocok", "s3cret", map[string]string{"Authorization": "Bearer s3cret"}, true},
		{"x-cron-secret cocok", "s3cret", map[string]string{"X-Cron-Secret": "s3cret"}, true},
		{"bearer salah, header cocok", "s3cret", map[string]string{"Authorization": "Bearer x", "X-Cron-Secret": "s3cret"}, true},
		{"bearer salah", "s3cret", map[string]string{"Authorization": "Bearer s3cre"}, false},
		{"tanpa prefix Bearer", "s3cret", map[string]string{"Authorization": "s3cret"}, false},
		{"tanpa header", "s3cret", nil, false},
		{"secret kosong", "", map[string]string{"Authorization": "Bearer ", "X-Cron-Secret": ""}, false},
	}
	for _, tc := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		for k, v := range tc.header {
			c.Request.Header.Set(k, v)
		}
		if got := CronAuthorized(c, tc.secret); got != tc.want {
			t.Errorf("%s: %v, ingin %v", tc.name, got, tc.want)
		}
	}
}
