package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/praktikum"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
)

func NewRouter(cfg *config.Config) (*gin.Engine, error) {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Vercel menulis X-Vercel-Forwarded-For di edge; client tidak bisa
	// mengirim header ini (Vercel menimpa/menghapusnya saat masuk).
	// Gin v1.10.0 belum punya gin.PlatformVercel, jadi pakai nama headernya
	// langsung — TrustedPlatform memang berupa nama header.
	// Combined dengan SetTrustedProxies(nil) di bawah: X-Forwarded-For/X-Real-IP
	// dari client selalu diabaikan.
	const PlatformVercel = "X-Vercel-Forwarded-For"
	r.TrustedPlatform = PlatformVercel
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.MaxMultipartMemory = 10 << 20 // 10 MiB: batasi buffer multipart di memori
	r.Use(
		middleware.CORS(middleware.NewOriginMatcher(cfg.CORSOrigins, cfg.CORSOriginPatterns)),
		gin.Logger(),
		gin.Recovery(),
	)

	r.GET("/api/health", health)
	if !cfg.IsProduction() {
		registerSwagger(r)
	}
	return r, nil
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "algohub-api"})
}

type PraktikumHandlers struct {
	Auth         *handler.AuthHandler
	Dashboard    *handler.DashboardHandler
	Sesi         *handler.SesiHandler
	Aktivasi     *handler.AktivasiHandler
	Soal         *handler.SoalHandler
	Jawaban      *handler.JawabanHandler
	Penilaian    *handler.PenilaianHandler
	Konfigurasi  *handler.KonfigurasiHandler
	Profile      *handler.ProfileHandler
	Jadwal       *handler.JadwalHandler
	User         *handler.UserHandler
	Kelas        *handler.KelasHandler
	Pedoman      *handler.PedomanHandler
	Praktikum    *handler.PraktikumHandler
	Upload       *handler.UploadHandler
	Ampuan       *handler.AmpuanHandler
	Rekap        *handler.RekapHandler
	AIGrading    *handler.AIGradingHandler
	RekapJawaban *handler.RekapJawabanHandler
	Cron         *handler.CronHandler
	Run          *handler.RunHandler
	AuditLog     *handler.AuditLogHandler
	Belajar      *handler.BelajarHandler
	Game         *handler.GameHandler
	LaporanBesar *handler.LaporanBesarHandler
}

type Praktikum struct {
	Config      *config.Config
	JWT         *jwt.Manager
	Users       repository.UserRepository
	Blocklist   repository.TokenBlocklist
	Konfigurasi repository.KonfigurasiRepository
	Handlers    PraktikumHandlers
}

// Middleware ber-state (rate limiter, auth) dibuat sekali dan dipakai bersama path baru dan alias,
// supaya alias tidak menggandakan kuota brute-force.
func RegisterPraktikum(r *gin.Engine, p *Praktikum) {
	h := p.Handlers
	// Rate-limit untuk meredam brute-force: maks 20 percobaan/menit per IP.
	authLimit := middleware.RateLimit(20, time.Minute)
	loginLimit := middleware.RateLimit(10, time.Minute)
	runLimit := middleware.RateLimitUser(30, time.Minute)
	authmw := middleware.PraktikumAuth(p.Config, p.JWT, p.Users, p.Blocklist)

	api := r.Group("/api")

	// Dipicu cron eksternal, dijaga X-Cron-Secret.
	api.POST("/cron/praktikum/auto-submit", h.Cron.AutoSubmit)
	api.POST("/cron/praktikum/rumpang", h.Cron.Rumpang)
	// ALIAS-LAMA: hapus setelah cutover
	api.POST("/cron/auto-submit", h.Cron.AutoSubmit)
	api.POST("/cron/rumpang", h.Cron.Rumpang)

	prak := api.Group("/praktikum", authmw, middleware.RequireRole(string(entity.RoleMahasiswa)))
	prak.GET("/dashboard", h.Praktikum.Dashboard)
	prak.GET("/sesi", h.Praktikum.ListSesi)
	prak.GET("/ruang", h.Jawaban.GetRuang)
	prak.POST("/mulai", h.Jawaban.Mulai)
	prak.POST("/autosave", h.Jawaban.AutoSave)
	prak.POST("/submit", h.Jawaban.Submit)
	prak.POST("/run", runLimit, h.Run.Execute)
	// Laporan Besar (tugas akhir praktikum). Berkas via signed URL ke bucket privat.
	prak.GET("/laporan-besar", h.LaporanBesar.StatusSaya)
	prak.POST("/laporan-besar/upload-url", h.LaporanBesar.UploadURL)
	prak.POST("/laporan-besar/kumpul", h.LaporanBesar.Kumpul)
	prak.DELETE("/laporan-besar/kumpul", h.LaporanBesar.Batalkan)

	mountPraktikum(api.Group("/praktikum"), p, authmw, authLimit, loginLimit)
	mountPraktikum(api, p, authmw, authLimit, loginLimit) // ALIAS-LAMA: hapus setelah cutover
}

func mountPraktikum(g *gin.RouterGroup, p *Praktikum, authmw, authLimit, loginLimit gin.HandlerFunc) {
	h := p.Handlers
	write := func(resource string) gin.HandlerFunc {
		return middleware.RequirePermission(p.Konfigurasi, resource, "write")
	}

	auth := g.Group("/auth", authLimit)
	auth.POST("/cek-nim", h.Auth.CekNIM)
	// LOW-04: login lebih ketat (10/menit/IP) — brute-force password langsung
	// dihambat, terpisah dari kuota endpoint auth lain.
	auth.POST("/login", loginLimit, h.Auth.Login)
	auth.POST("/register", h.Auth.Register)
	auth.POST("/forgot-password", h.Auth.ForgotPassword)
	auth.POST("/reset-password", h.Auth.ResetPassword)

	info := g.Group("/info")
	info.GET("/asisten", h.User.ListAsisten)
	info.GET("/laporan", h.Pedoman.List)
	info.GET("/modul", h.Konfigurasi.PublicModul)
	info.GET("/jadwal", h.Jadwal.List)
	info.GET("/jadwal/config", h.Konfigurasi.PublicJadwalConfig)
	info.GET("/kelas/:id/mahasiswa", h.Ampuan.PublicKelasMahasiswa)
	info.GET("/announcements", h.Konfigurasi.PublicAnnouncements)

	authed := g.Group("", authmw)
	authed.POST("/auth/logout", h.Auth.Logout)
	authed.GET("/auth/me", h.Auth.Me)
	authed.GET("/profile", h.Profile.Get)
	authed.PUT("/profile", h.Profile.Update)
	// Belajar (elearning). Materi di sini TANPA kunci jawaban.
	authed.GET("/belajar/level", h.Belajar.ListLevel)
	authed.GET("/belajar/modul", h.Belajar.ListModul)
	authed.GET("/belajar/materi", h.Belajar.ListMateri)
	authed.GET("/belajar/materi/:id", h.Belajar.GetMateri)
	authed.GET("/belajar/progres", h.Belajar.ProgresSaya)
	authed.POST("/belajar/materi/:id/selesai", h.Belajar.SelesaikanMateri)
	authed.GET("/belajar/pencapaian", h.Belajar.PencapaianSaya)
	authed.GET("/belajar/sesi", h.Belajar.SesiBelajar)
	// Game bug-hunt. Soal dikirim tanpa kunci; jawaban dinilai server.
	authed.GET("/game/konfigurasi", h.Game.Konfigurasi)
	authed.GET("/game/status", h.Game.StatusMain)
	authed.POST("/game/mulai", h.Game.MulaiMain)
	authed.POST("/game/selesai", h.Game.SelesaiMain)
	authed.GET("/game/peringkat", h.Game.Peringkat)
	authed.POST("/game/detak", h.Game.Detak)

	admin := g.Group("/admin", authmw, middleware.RequireRole(string(entity.RoleAsisten), string(entity.RoleKoordinator)))
	admin.GET("/dashboard", h.Dashboard.Statistik)

	admin.GET("/users", h.User.ListMahasiswa)
	admin.POST("/users", write("users"), h.User.CreateMahasiswa)
	admin.POST("/users/bulk", write("users"), h.User.BulkUpsertMahasiswa)
	admin.PUT("/users/bulk-attrs", write("users"), h.User.BulkAttrsMahasiswa)
	admin.PUT("/users/:id", write("users"), h.User.UpdateMahasiswa)
	admin.DELETE("/users/:id", write("users"), h.User.Delete)
	admin.POST("/users/:id/reset-password", write("users"), h.User.ResetPassword)

	// Materi LENGKAP dengan kunci jawaban, jangan dipakai frontend mahasiswa.
	admin.GET("/materi", h.Belajar.AdminListMateri)
	admin.GET("/materi/:id", h.Belajar.AdminGetMateri)
	admin.POST("/materi", write("sesi"), h.Belajar.AdminCreateMateri)
	admin.PUT("/materi/:id", write("sesi"), h.Belajar.AdminUpdateMateri)
	admin.DELETE("/materi/:id", write("sesi"), h.Belajar.AdminDeleteMateri)

	admin.PUT("/belajar/kurikulum", write("sesi"), h.Belajar.AdminSimpanKurikulum)
	admin.DELETE("/belajar/kurikulum", write("sesi"), h.Belajar.AdminKosongkanKurikulum)
	admin.GET("/belajar/users", h.Belajar.AdminListUserBelajar)
	admin.GET("/belajar/users/:id/progres", h.Belajar.AdminProgresUser)
	admin.PUT("/belajar/users/:id/akses-level", write("penilaian"), h.Belajar.AdminSetAksesLevel)
	admin.POST("/belajar/users/:id/reset", write("penilaian"), h.Belajar.AdminResetProgres)
	admin.PUT("/belajar/users/:id/xp", write("penilaian"), h.Belajar.AdminSetXP)

	admin.GET("/monitoring/sesi", h.Game.SesiAktif)
	admin.GET("/game/soal", h.Game.AdminListSoal)
	admin.POST("/game/soal", write("sesi"), h.Game.AdminCreateSoal)
	admin.PUT("/game/soal/:id", write("sesi"), h.Game.AdminUpdateSoal)
	admin.DELETE("/game/soal/:id", write("sesi"), h.Game.AdminDeleteSoal)
	admin.PUT("/game/konfigurasi", write("konfigurasi"), h.Game.AdminSimpanKonfigurasi)

	admin.GET("/kelas", h.Kelas.List)
	admin.POST("/kelas", write("kelas"), h.Kelas.Create)
	admin.PUT("/kelas/:id", write("kelas"), h.Kelas.Update)
	admin.DELETE("/kelas/:id", write("kelas"), h.Kelas.Delete)

	admin.GET("/asisten", h.User.ListAsisten)
	admin.POST("/asisten", write("asisten"), h.User.CreateAsisten)
	admin.PUT("/asisten/:id", write("asisten"), h.User.UpdateAsisten)

	admin.GET("/jadwal", h.Jadwal.List)
	admin.POST("/jadwal", write("jadwal"), h.Jadwal.Create)
	admin.PUT("/jadwal/:id", write("jadwal"), h.Jadwal.Update)
	admin.DELETE("/jadwal/:id", write("jadwal"), h.Jadwal.Delete)

	admin.GET("/pedoman", h.Pedoman.List)
	admin.POST("/pedoman", write("konfigurasi"), h.Pedoman.Create)
	admin.PUT("/pedoman/:id", write("konfigurasi"), h.Pedoman.Update)
	admin.DELETE("/pedoman/:id", write("konfigurasi"), h.Pedoman.Delete)

	admin.GET("/konfigurasi", h.Konfigurasi.All)
	admin.GET("/konfigurasi/ai-model", h.Konfigurasi.AIModel)
	admin.POST("/konfigurasi", write("konfigurasi"), h.Konfigurasi.Set)

	admin.GET("/sesi", h.Sesi.List)
	admin.GET("/sesi/:id", h.Sesi.Get)
	admin.POST("/sesi", write("sesi"), h.Sesi.Create)
	admin.PUT("/sesi/:id", write("sesi"), h.Sesi.Update)
	admin.DELETE("/sesi/:id", write("sesi"), h.Sesi.Delete)
	admin.GET("/sesi/:id/course", h.Sesi.ListCourse)
	admin.POST("/sesi/:id/course", write("sesi"), h.Sesi.CreateCourse)
	admin.PUT("/course/:courseId", write("sesi"), h.Sesi.UpdateCourse)
	admin.DELETE("/course/:courseId", write("sesi"), h.Sesi.DeleteCourse)

	admin.GET("/soal", h.Soal.ListByCourse)
	admin.POST("/soal", write("sesi"), h.Soal.Create)
	admin.PUT("/soal/:id", write("sesi"), h.Soal.Update)
	admin.DELETE("/soal/:id", write("sesi"), h.Soal.Delete)

	admin.GET("/aktivasi", h.Aktivasi.List)
	admin.POST("/aktivasi", write("aktivasi"), h.Aktivasi.Aktivasi)
	admin.GET("/aktivasi/:id", h.Aktivasi.Get)
	admin.DELETE("/aktivasi/:id", write("aktivasi"), h.Aktivasi.Delete)
	admin.POST("/aktivasi/:id/token", write("aktivasi"), h.Aktivasi.GenerateToken)
	admin.POST("/aktivasi/:id/susulan", write("aktivasi"), h.Aktivasi.AddSusulan)
	admin.GET("/aktivasi/:id/susulan", h.Aktivasi.ListSusulan)
	admin.GET("/aktivasi/:id/peserta", h.Aktivasi.Peserta)
	admin.DELETE("/aktivasi/:id/susulan/:mahasiswaId", write("aktivasi"), h.Aktivasi.RemoveSusulan)
	admin.POST("/aktivasi-course/buka-tutup", write("aktivasi"), h.Aktivasi.BukaTutupCourse)
	admin.POST("/aktivasi-course/buka-tutup-peserta", write("aktivasi"), h.Aktivasi.BukaTutupPeserta)

	admin.GET("/penilaian/rekap", h.Penilaian.Rekap)
	admin.POST("/penilaian", write("penilaian"), h.Penilaian.SetNilai)
	admin.POST("/keaktifan", write("penilaian"), h.Penilaian.SetKeaktifan)
	// AI grading SINKRON 1-per-1 (frontend yang loop).
	admin.GET("/penilaian/ai-grade/targets", h.AIGrading.ListTargets)
	admin.POST("/penilaian/ai-grade/one", write("penilaian"), h.AIGrading.GradeOne)

	admin.GET("/rekap-jawaban", h.RekapJawaban.GetRekapJawabanGlobal)
	admin.POST("/penilaian/bulk-action", write("penilaian"), h.RekapJawaban.BulkAction)
	admin.POST("/jawaban/inject", write("penilaian"), h.Jawaban.AdminInjectJawaban)

	admin.GET("/rekap/kelas/:id_kelas", h.Rekap.GetRekapKelas)

	admin.POST("/upload", write("sesi"), h.Upload.Upload)

	admin.GET("/ampuan", h.Ampuan.List)
	admin.POST("/ampuan", write("asisten"), h.Ampuan.Create)
	admin.DELETE("/ampuan/:id", write("asisten"), h.Ampuan.Delete)

	admin.GET("/audit-logs", h.AuditLog.GetLogs)

	// Laporan Besar (tugas akhir). Tulis butuh permission "sesi", baca cukup role admin.
	admin.GET("/laporan-besar", h.LaporanBesar.List)
	admin.POST("/laporan-besar", write("sesi"), h.LaporanBesar.Create)
	admin.GET("/laporan-besar/:id", h.LaporanBesar.GetDetail)
	admin.PUT("/laporan-besar/:id", write("sesi"), h.LaporanBesar.Update)
	admin.DELETE("/laporan-besar/:id", write("sesi"), h.LaporanBesar.Delete)
	admin.PUT("/laporan-besar/:id/target", write("sesi"), h.LaporanBesar.SetTarget)
	admin.GET("/laporan-besar/:id/pengumpulan", h.LaporanBesar.ListPengumpulan)
	admin.PUT("/laporan-besar/:id/pengumpulan/:mhsId/mark", write("sesi"), h.LaporanBesar.Mark)
	admin.PUT("/laporan-besar/:id/pengumpulan/:mhsId/override", write("sesi"), h.LaporanBesar.SetOverride)
	admin.DELETE("/laporan-besar/:id/pengumpulan/:mhsId/override", write("sesi"), h.LaporanBesar.RevokeOverride)
	admin.GET("/laporan-besar/:id/pengumpulan/:mhsId/preview", h.LaporanBesar.Preview)
	admin.GET("/laporan-besar/:id/pengumpulan/:mhsId/download", h.LaporanBesar.Download)
	admin.POST("/laporan-besar/:id/download-batch", write("sesi"), h.LaporanBesar.DownloadBatch)

	superadmin := g.Group("/superadmin", authmw, middleware.RequireRole(string(entity.RoleKoordinator)))
	superadmin.GET("/permissions", h.Konfigurasi.GetRolePermissions)
	superadmin.PUT("/permissions", h.Konfigurasi.SetRolePermissions)
}
