package handler

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

// Body yang gagal di-parse diperlakukan sebagai {} seperti `req.body || {}` di kode lama,
// sehingga pesan validasi endpoint itu sendiri yang keluar.
func bindBody[T any](c *gin.Context) T {
	var v T
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil || json.Unmarshal(raw, &v) != nil {
		var zero T
		return zero
	}
	return v
}
