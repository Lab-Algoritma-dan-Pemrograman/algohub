package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
)

const CtxElearningPrincipal = "elearning_principal"

// Pesan 401 sama dengan api/ai.ts dan api/grade.ts lama.
func ElearningAuth(a interface {
	Authenticate(string) (*dto.Principal, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || token == "" {
			response.Fail(c, http.StatusUnauthorized, "Unauthorized: Missing token", nil)
			return
		}
		p, err := a.Authenticate(token)
		if err != nil || p == nil {
			response.Fail(c, http.StatusUnauthorized, "Unauthorized: Invalid token", nil)
			return
		}
		c.Set(CtxElearningPrincipal, p)
		c.Next()
	}
}

func ElearningPrincipal(c *gin.Context) *dto.Principal {
	if v, ok := c.Get(CtxElearningPrincipal); ok {
		if p, ok := v.(*dto.Principal); ok {
			return p
		}
	}
	return nil
}
