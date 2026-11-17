package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
)

type routeKey struct {
	method string
	path   string
}

// Disalin dari router praktikum lama (tanpa /api/health, /swagger, dan compile-c).
var legacyRoutes = []routeKey{
	{http.MethodPost, "/api/cron/auto-submit"},
	{http.MethodPost, "/api/cron/rumpang"},
	{http.MethodPost, "/api/auth/cek-nim"},
	{http.MethodPost, "/api/auth/login"},
	{http.MethodPost, "/api/auth/register"},
	{http.MethodPost, "/api/auth/forgot-password"},
	{http.MethodPost, "/api/auth/reset-password"},
	{http.MethodGet, "/api/info/asisten"},
	{http.MethodGet, "/api/info/laporan"},
	{http.MethodGet, "/api/info/modul"},
	{http.MethodGet, "/api/info/jadwal"},
	{http.MethodGet, "/api/info/jadwal/config"},
	{http.MethodGet, "/api/info/kelas/:id/mahasiswa"},
	{http.MethodGet, "/api/info/announcements"},
	{http.MethodPost, "/api/auth/logout"},
	{http.MethodGet, "/api/auth/me"},
	{http.MethodGet, "/api/profile"},
	{http.MethodPut, "/api/profile"},
	{http.MethodGet, "/api/belajar/level"},
	{http.MethodGet, "/api/belajar/modul"},
	{http.MethodGet, "/api/belajar/materi"},
	{http.MethodGet, "/api/belajar/materi/:id"},
	{http.MethodGet, "/api/belajar/progres"},
	{http.MethodPost, "/api/belajar/materi/:id/selesai"},
	{http.MethodGet, "/api/belajar/pencapaian"},
	{http.MethodGet, "/api/belajar/sesi"},
	{http.MethodGet, "/api/game/konfigurasi"},
	{http.MethodGet, "/api/game/status"},
	{http.MethodPost, "/api/game/mulai"},
	{http.MethodPost, "/api/game/selesai"},
	{http.MethodGet, "/api/game/peringkat"},
	{http.MethodPost, "/api/game/detak"},
	{http.MethodGet, "/api/praktikum/dashboard"},
	{http.MethodGet, "/api/praktikum/sesi"},
	{http.MethodGet, "/api/praktikum/ruang"},
	{http.MethodPost, "/api/praktikum/mulai"},
	{http.MethodPost, "/api/praktikum/autosave"},
	{http.MethodPost, "/api/praktikum/submit"},
	{http.MethodPost, "/api/praktikum/run"},
	{http.MethodGet, "/api/praktikum/laporan-besar"},
	{http.MethodPost, "/api/praktikum/laporan-besar/upload-url"},
	{http.MethodPost, "/api/praktikum/laporan-besar/kumpul"},
	{http.MethodDelete, "/api/praktikum/laporan-besar/kumpul"},
	{http.MethodGet, "/api/admin/dashboard"},
	{http.MethodGet, "/api/admin/users"},
	{http.MethodPost, "/api/admin/users"},
	{http.MethodPost, "/api/admin/users/bulk"},
	{http.MethodPut, "/api/admin/users/bulk-attrs"},
	{http.MethodPut, "/api/admin/users/:id"},
	{http.MethodDelete, "/api/admin/users/:id"},
	{http.MethodPost, "/api/admin/users/:id/reset-password"},
	{http.MethodGet, "/api/admin/materi"},
	{http.MethodGet, "/api/admin/materi/:id"},
	{http.MethodPost, "/api/admin/materi"},
	{http.MethodPut, "/api/admin/materi/:id"},
	{http.MethodDelete, "/api/admin/materi/:id"},
	{http.MethodPut, "/api/admin/belajar/kurikulum"},
	{http.MethodDelete, "/api/admin/belajar/kurikulum"},
	{http.MethodGet, "/api/admin/belajar/users"},
	{http.MethodGet, "/api/admin/belajar/users/:id/progres"},
	{http.MethodPut, "/api/admin/belajar/users/:id/akses-level"},
	{http.MethodPost, "/api/admin/belajar/users/:id/reset"},
	{http.MethodPut, "/api/admin/belajar/users/:id/xp"},
	{http.MethodGet, "/api/admin/monitoring/sesi"},
	{http.MethodGet, "/api/admin/game/soal"},
	{http.MethodPost, "/api/admin/game/soal"},
	{http.MethodPut, "/api/admin/game/soal/:id"},
	{http.MethodDelete, "/api/admin/game/soal/:id"},
	{http.MethodPut, "/api/admin/game/konfigurasi"},
	{http.MethodGet, "/api/admin/kelas"},
	{http.MethodPost, "/api/admin/kelas"},
	{http.MethodPut, "/api/admin/kelas/:id"},
	{http.MethodDelete, "/api/admin/kelas/:id"},
	{http.MethodGet, "/api/admin/asisten"},
	{http.MethodPost, "/api/admin/asisten"},
	{http.MethodPut, "/api/admin/asisten/:id"},
	{http.MethodGet, "/api/admin/jadwal"},
	{http.MethodPost, "/api/admin/jadwal"},
	{http.MethodPut, "/api/admin/jadwal/:id"},
	{http.MethodDelete, "/api/admin/jadwal/:id"},
	{http.MethodGet, "/api/admin/pedoman"},
	{http.MethodPost, "/api/admin/pedoman"},
	{http.MethodPut, "/api/admin/pedoman/:id"},
	{http.MethodDelete, "/api/admin/pedoman/:id"},
	{http.MethodGet, "/api/admin/konfigurasi"},
	{http.MethodGet, "/api/admin/konfigurasi/ai-model"},
	{http.MethodPost, "/api/admin/konfigurasi"},
	{http.MethodGet, "/api/admin/sesi"},
	{http.MethodGet, "/api/admin/sesi/:id"},
	{http.MethodPost, "/api/admin/sesi"},
	{http.MethodPut, "/api/admin/sesi/:id"},
	{http.MethodDelete, "/api/admin/sesi/:id"},
	{http.MethodGet, "/api/admin/sesi/:id/course"},
	{http.MethodPost, "/api/admin/sesi/:id/course"},
	{http.MethodPut, "/api/admin/course/:courseId"},
	{http.MethodDelete, "/api/admin/course/:courseId"},
	{http.MethodGet, "/api/admin/soal"},
	{http.MethodPost, "/api/admin/soal"},
	{http.MethodPut, "/api/admin/soal/:id"},
	{http.MethodDelete, "/api/admin/soal/:id"},
	{http.MethodGet, "/api/admin/aktivasi"},
	{http.MethodPost, "/api/admin/aktivasi"},
	{http.MethodGet, "/api/admin/aktivasi/:id"},
	{http.MethodDelete, "/api/admin/aktivasi/:id"},
	{http.MethodPost, "/api/admin/aktivasi/:id/token"},
	{http.MethodPost, "/api/admin/aktivasi/:id/susulan"},
	{http.MethodGet, "/api/admin/aktivasi/:id/susulan"},
	{http.MethodGet, "/api/admin/aktivasi/:id/peserta"},
	{http.MethodDelete, "/api/admin/aktivasi/:id/susulan/:mahasiswaId"},
	{http.MethodPost, "/api/admin/aktivasi-course/buka-tutup"},
	{http.MethodPost, "/api/admin/aktivasi-course/buka-tutup-peserta"},
	{http.MethodGet, "/api/admin/penilaian/rekap"},
	{http.MethodPost, "/api/admin/penilaian"},
	{http.MethodPost, "/api/admin/keaktifan"},
	{http.MethodGet, "/api/admin/penilaian/ai-grade/targets"},
	{http.MethodPost, "/api/admin/penilaian/ai-grade/one"},
	{http.MethodGet, "/api/admin/rekap-jawaban"},
	{http.MethodPost, "/api/admin/penilaian/bulk-action"},
	{http.MethodPost, "/api/admin/jawaban/inject"},
	{http.MethodGet, "/api/admin/rekap/kelas/:id_kelas"},
	{http.MethodPost, "/api/admin/upload"},
	{http.MethodGet, "/api/admin/ampuan"},
	{http.MethodPost, "/api/admin/ampuan"},
	{http.MethodDelete, "/api/admin/ampuan/:id"},
	{http.MethodGet, "/api/admin/audit-logs"},
	{http.MethodGet, "/api/admin/laporan-besar"},
	{http.MethodPost, "/api/admin/laporan-besar"},
	{http.MethodGet, "/api/admin/laporan-besar/:id"},
	{http.MethodPut, "/api/admin/laporan-besar/:id"},
	{http.MethodDelete, "/api/admin/laporan-besar/:id"},
	{http.MethodPut, "/api/admin/laporan-besar/:id/target"},
	{http.MethodGet, "/api/admin/laporan-besar/:id/pengumpulan"},
	{http.MethodPut, "/api/admin/laporan-besar/:id/pengumpulan/:mhsId/mark"},
	{http.MethodPut, "/api/admin/laporan-besar/:id/pengumpulan/:mhsId/override"},
	{http.MethodDelete, "/api/admin/laporan-besar/:id/pengumpulan/:mhsId/override"},
	{http.MethodGet, "/api/admin/laporan-besar/:id/pengumpulan/:mhsId/preview"},
	{http.MethodGet, "/api/admin/laporan-besar/:id/pengumpulan/:mhsId/download"},
	{http.MethodPost, "/api/admin/laporan-besar/:id/download-batch"},
	{http.MethodGet, "/api/superadmin/permissions"},
	{http.MethodPut, "/api/superadmin/permissions"},
}

func newPath(old string) string {
	switch {
	case strings.HasPrefix(old, "/api/cron/"):
		return "/api/cron/praktikum/" + strings.TrimPrefix(old, "/api/cron/")
	case strings.HasPrefix(old, "/api/praktikum/"):
		return old
	default:
		return "/api/praktikum/" + strings.TrimPrefix(old, "/api/")
	}
}

func newPraktikumRouter(t *testing.T, appEnv string) *gin.Engine {
	t.Helper()
	mode := gin.Mode()
	t.Cleanup(func() { gin.SetMode(mode) })
	gin.SetMode(gin.TestMode)
	r, err := NewRouter(&config.Config{AppEnv: appEnv, CORSOrigins: []string{"https://praktikum.algohub.web.id"}})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	RegisterPraktikum(r, &Praktikum{
		Config:   &config.Config{AppEnv: appEnv},
		JWT:      jwt.NewManager("uji", 1),
		Handlers: PraktikumHandlers{Cron: handler.NewCronHandler(nil, nil, "")},
	})
	return r
}

func routeHandlers(r *gin.Engine) map[routeKey]string {
	m := make(map[routeKey]string)
	for _, ri := range r.Routes() {
		m[routeKey{ri.Method, ri.Path}] = ri.Handler
	}
	return m
}

func TestPraktikumRoutesNewAndAlias(t *testing.T) {
	r := newPraktikumRouter(t, "test")
	routes := routeHandlers(r)

	if len(legacyRoutes) != 140 {
		t.Fatalf("legacyRoutes = %d, want 140", len(legacyRoutes))
	}
	aliases := 0
	for _, old := range legacyRoutes {
		nk := routeKey{old.method, newPath(old.path)}
		nh, ok := routes[nk]
		if !ok {
			t.Errorf("route baru %s %s tidak terdaftar", nk.method, nk.path)
			continue
		}
		if nk.path == old.path {
			continue
		}
		aliases++
		oh, ok := routes[old]
		if !ok {
			t.Errorf("alias %s %s tidak terdaftar", old.method, old.path)
			continue
		}
		if oh != nh {
			t.Errorf("alias %s %s handler %q, route baru %q", old.method, old.path, oh, nh)
		}
	}
	if aliases != 129 {
		t.Errorf("alias = %d, want 129", aliases)
	}
	if got := len(r.Routes()); got != 271 {
		t.Errorf("total route = %d, want 271 (health + swagger + 140 + 129)", got)
	}
}

func TestPraktikumRoutesNoCompileOrDoublePrefix(t *testing.T) {
	r := newPraktikumRouter(t, "test")
	for _, ri := range r.Routes() {
		if strings.HasSuffix(ri.Path, "/compile-c") {
			t.Errorf("route compile-c masih terdaftar: %s %s", ri.Method, ri.Path)
		}
		if strings.HasPrefix(ri.Path, "/api/praktikum/praktikum/") {
			t.Errorf("prefix ganda: %s %s", ri.Method, ri.Path)
		}
		if strings.HasPrefix(ri.Path, "/api/praktikum/cron/") {
			t.Errorf("cron di bawah /api/praktikum: %s %s", ri.Method, ri.Path)
		}
	}
	if w := serve(r, http.MethodPost, "/api/praktikum/compile-c", "", nil); w.Code != http.StatusNotFound {
		t.Errorf("POST /api/praktikum/compile-c = %d, want 404", w.Code)
	}
}

func TestPraktikumSwaggerOnlyOutsideProduction(t *testing.T) {
	has := func(r *gin.Engine) bool {
		for _, ri := range r.Routes() {
			if ri.Path == "/swagger/*any" {
				return true
			}
		}
		return false
	}
	if !has(newPraktikumRouter(t, "test")) {
		t.Error("swagger tidak terpasang di non-production")
	}
	prod := newPraktikumRouter(t, "production")
	if has(prod) {
		t.Error("swagger terpasang di production")
	}
	if got := len(prod.Routes()); got != 270 {
		t.Errorf("total route production = %d, want 270", got)
	}
}

func TestPraktikumAuthChain(t *testing.T) {
	r := newPraktikumRouter(t, "test")
	cases := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/api/praktikum/admin/dashboard", http.StatusUnauthorized},
		{http.MethodGet, "/api/admin/dashboard", http.StatusUnauthorized},
		{http.MethodGet, "/api/praktikum/dashboard", http.StatusUnauthorized},
		{http.MethodGet, "/api/praktikum/superadmin/permissions", http.StatusUnauthorized},
		{http.MethodPost, "/api/praktikum/cron/auto-submit", http.StatusNotFound},
		{http.MethodPost, "/api/cron/praktikum/auto-submit", http.StatusUnauthorized},
		{http.MethodPost, "/api/cron/praktikum/rumpang", http.StatusUnauthorized},
		{http.MethodPost, "/api/cron/auto-submit", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		if w := serve(r, tc.method, tc.path, "", nil); w.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, w.Code, tc.want)
		}
	}
}

func TestPraktikumLoginLimitSharedWithAlias(t *testing.T) {
	r := newPraktikumRouter(t, "test")
	ip := map[string]string{"X-Forwarded-For": "203.0.113.7"}
	for i := range 10 {
		if w := serve(r, http.MethodPost, "/api/praktikum/auth/login", "", ip); w.Code != http.StatusBadRequest {
			t.Fatalf("login ke-%d = %d, want 400", i+1, w.Code)
		}
	}
	if w := serve(r, http.MethodPost, "/api/auth/login", "", ip); w.Code != http.StatusTooManyRequests {
		t.Errorf("alias login setelah kuota habis = %d, want 429", w.Code)
	}
	other := map[string]string{"X-Forwarded-For": "203.0.113.8"}
	if w := serve(r, http.MethodPost, "/api/auth/login", "", other); w.Code != http.StatusBadRequest {
		t.Errorf("login dari IP lain = %d, want 400", w.Code)
	}
}

var _ = httptest.NewRecorder
