package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
	supajwt "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/supabasejwt"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID      = "user_id"
	CtxNIM         = "nim"
	CtxRole        = "role"
	CtxTokenID     = "token_jti"
	CtxTokenExpiry = "token_exp"
)

// PraktikumAuth memverifikasi JWT dual-mode: legacy HS256 lokal dan/atau Supabase RS256/ES256.
// Role untuk Supabase SELALU resolve dari DB via supabase_user_id (tidak dari claim).
// block opsional: bila disediakan, token dengan jti yang sudah logout ditolak (401).
func PraktikumAuth(cfg *config.Config, jm *jwt.Manager, userRepo repository.UserRepository, block ...repository.TokenBlocklist) gin.HandlerFunc {
	var blocklist repository.TokenBlocklist
	if len(block) > 0 {
		blocklist = block[0]
	}
	mode := strings.ToLower(strings.TrimSpace(cfg.AuthMode))
	if mode == "" {
		mode = "legacy"
	}
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "Token tidak ditemukan", nil)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if mode == "legacy" {
			if legacyAuth(c, jm, tokenStr, userRepo, blocklist) {
				c.Next()
				return
			}
			response.Fail(c, http.StatusUnauthorized, "Token tidak valid / kedaluwarsa", nil)
			return
		}
		if mode == "supabase" {
			if supabaseAuth(c, cfg, userRepo, tokenStr) {
				c.Next()
				return
			}
			response.Fail(c, http.StatusUnauthorized, "Token tidak valid / kedaluwarsa", nil)
			return
		}
		// dual: coba Supabase dulu lalu legacy
		if supabaseAuth(c, cfg, userRepo, tokenStr) {
			c.Next()
			return
		}
		if legacyAuth(c, jm, tokenStr, userRepo, blocklist) {
			c.Next()
			return
		}
		response.Fail(c, http.StatusUnauthorized, "Token tidak valid / kedaluwarsa", nil)
	}
}

func legacyAuth(c *gin.Context, jm *jwt.Manager, tokenStr string, userRepo repository.UserRepository, blocklist repository.TokenBlocklist) bool {
	claims, err := jm.Verify(tokenStr)
	if err != nil {
		return false
	}
	// K-1: token yang sudah logout (jti ada di blocklist) ditolak. Token legacy
	// tanpa jti lolos — jangan tolak semua token lama yang masih beredar.
	if blocklist != nil && claims.ID != "" {
		revoked, err := blocklist.IsTokenRevoked(c.Request.Context(), claims.ID)
		if err == nil && revoked {
			return false
		}
		// err != nil: fail-open, sama seperti keputusan siakad.
	}
	// CR-M3: identitas & ROLE di-resolve ulang dari DB, bukan dari claim.
	// Dulu claims.Role dipakai apa adanya, jadi user yang diturunkan/dihapus
	// tetap memegang akses admin sampai token 12 jam habis.
	if userRepo == nil {
		return false
	}
	u, err := userRepo.FindByID(claims.UserID)
	if err != nil || u == nil {
		return false
	}
	c.Set(CtxUserID, u.ID)
	c.Set(CtxNIM, u.NIM)
	c.Set(CtxRole, string(u.Role))
	if claims.ID != "" {
		c.Set(CtxTokenID, claims.ID)
	}
	if exp, ok := jwt.ExpiryTime(claims); ok {
		c.Set(CtxTokenExpiry, exp)
	}
	return true
}

func supabaseAuth(c *gin.Context, cfg *config.Config, userRepo repository.UserRepository, tokenStr string) bool {
	claims, err := supajwt.VerifySupabaseToken(tokenStr, cfg.SupabaseJWKSURL, cfg.SupabaseJWTIssuer, cfg.SupabaseJWTAud)
	if err != nil {
		return false
	}
	if userRepo == nil {
		return false
	}
	u, err := userRepo.FindBySupabaseUserID(claims.Subject)
	if err != nil || u == nil {
		return false
	}
	c.Set(CtxUserID, u.ID)
	c.Set(CtxNIM, u.NIM)
	c.Set(CtxRole, string(u.Role))
	return true
}

func UserID(c *gin.Context) int {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(int); ok {
			return id
		}
	}
	return 0
}

// TokenID: jti token praktikum; "" untuk token legacy / Supabase.
func TokenID(c *gin.Context) string {
	if v, ok := c.Get(CtxTokenID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// TokenExpiry: exp klaim token praktikum; false bila tak tersedia.
func TokenExpiry(c *gin.Context) (time.Time, bool) {
	if v, ok := c.Get(CtxTokenExpiry); ok {
		if t, ok := v.(time.Time); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

func Role(c *gin.Context) string {
	if v, ok := c.Get(CtxRole); ok {
		if r, ok := v.(string); ok {
			return r
		}
	}
	return ""
}
