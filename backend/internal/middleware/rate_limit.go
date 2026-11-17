package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Fixed window in-memory: melindungi endpoint sensitif (login/register/cek-nim) dari
// brute-force dan endpoint mahal (run) dari penyalahgunaan.
type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string]*window
	limit  int
	window time.Duration
}

type window struct {
	count int
	reset time.Time
}

func RateLimit(limit int, win time.Duration) gin.HandlerFunc {
	return newRateLimiter(limit, win, func(c *gin.Context) string {
		return "ip:" + ClientIP(c)
	})
}

// RateLimitUser membatasi per USER (JWT), fallback ke IP bila belum terautentikasi.
// Dipakai untuk /praktikum/run: endpoint ini meneruskan eksekusi kode ke Glot.io
// dengan token server, jadi satu akun tidak boleh menghabiskan kuota bersama.
func RateLimitUser(limit int, win time.Duration) gin.HandlerFunc {
	return newRateLimiter(limit, win, func(c *gin.Context) string {
		if uid := UserID(c); uid > 0 {
			return "user:" + strconv.Itoa(uid)
		}
		return "ip:" + ClientIP(c)
	})
}

func newRateLimiter(limit int, win time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	rl := &rateLimiter{
		hits:   make(map[string]*window),
		limit:  limit,
		window: win,
	}
	rl.startJanitor()

	return func(c *gin.Context) {
		if !rl.allow(keyFn(c)) {
			response.Fail(c, http.StatusTooManyRequests,
				"Terlalu banyak permintaan. Coba lagi beberapa saat.", nil)
			return
		}
		c.Next()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	w, ok := rl.hits[key]
	if !ok || now.After(w.reset) {
		rl.hits[key] = &window{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.limit {
		return false
	}
	w.count++
	return true
}

// startJanitor membersihkan entri kedaluwarsa secara berkala agar map tidak tumbuh tanpa batas.
func (rl *rateLimiter) startJanitor() {
	go func() {
		t := time.NewTicker(rl.window)
		defer t.Stop()
		for range t.C {
			rl.mu.Lock()
			now := time.Now()
			for k, w := range rl.hits {
				if now.After(w.reset) {
					delete(rl.hits, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
