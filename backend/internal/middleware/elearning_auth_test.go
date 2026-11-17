package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
)

type fakeAuthenticator map[string]*dto.Principal

func (f fakeAuthenticator) Authenticate(token string) (*dto.Principal, error) {
	if p, ok := f[token]; ok {
		return p, nil
	}
	return nil, errors.New("token tidak dikenal")
}

func TestElearningAuth(t *testing.T) {
	auth := fakeAuthenticator{"baik": {NIM: "202711001"}}
	var got *dto.Principal
	r := limitedEngine(t, ElearningAuth(auth), func(c *gin.Context) {
		got = ElearningPrincipal(c)
		c.Next()
	})

	cases := []struct {
		header map[string]string
		want   int
	}{
		{nil, http.StatusUnauthorized},
		{map[string]string{"Authorization": "Bearer "}, http.StatusUnauthorized},
		{map[string]string{"Authorization": "baik"}, http.StatusUnauthorized},
		{map[string]string{"Authorization": "Bearer rusak"}, http.StatusUnauthorized},
		{map[string]string{"Authorization": "Bearer baik"}, http.StatusOK},
	}
	for _, tc := range cases {
		if code := hit(r, tc.header); code != tc.want {
			t.Errorf("header %v = %d, want %d", tc.header, code, tc.want)
		}
	}
	if got == nil || got.NIM != "202711001" {
		t.Errorf("principal = %+v", got)
	}
}
