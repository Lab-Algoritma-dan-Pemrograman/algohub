package app

import (
	"gorm.io/gorm"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/praktikum"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/glot"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/hash"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/ollama"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/supabase"
)

func fbScryptConfig(cfg *config.Config) hash.FbScryptConfig {
	return hash.FbScryptConfig{
		SignerKey:     cfg.FbScryptSignerKey,
		SaltSeparator: cfg.FbScryptSaltSeparator,
		Rounds:        cfg.FbScryptRounds,
		MemCost:       cfg.FbScryptMemCost,
	}
}

func newPraktikum(cfg *config.Config, db *gorm.DB) *deliveryhttp.Praktikum {
	jm := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpireHours)
	sb := supabase.New(cfg.SupabaseURL, cfg.SupabaseServiceKey, cfg.SupabaseBucket)
	// Bucket privat khusus berkas laporan besar (baca via signed URL).
	sbLaporan := supabase.New(cfg.SupabaseURL, cfg.SupabaseServiceKey, cfg.SupabaseBucketLaporan)
	oc := ollama.NewClient(ollama.Config{URL: cfg.OllamaURL, Model: cfg.OllamaModel, APIKey: cfg.OllamaAPIKey})
	glotClient := glot.New(cfg.GlotURL, cfg.GlotToken)

	userRepo := repository.NewUserRepository(db)
	kelasRepo := repository.NewKelasRepository(db)
	jadwalRepo := repository.NewJadwalRepository(db)
	konfRepo := repository.NewKonfigurasiRepository(db)
	pedomanRepo := repository.NewPedomanRepository(db)
	sesiRepo := repository.NewSesiRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	soalRepo := repository.NewSoalRepository(db)
	aktivasiRepo := repository.NewAktivasiRepository(db)
	terpilihRepo := repository.NewSoalTerpilihRepository(db)
	jawabanRepo := repository.NewJawabanRepository(db)
	pengerjaanRepo := repository.NewPengerjaanRepository(db)
	ampuanRepo := repository.NewAmpuanRepository(db)
	rekapRepo := repository.NewRekapRepository(db)
	penilaianTxRepo := repository.NewPenilaianTxRepo(db)
	aktivasiTxRepo := repository.NewAktivasiTxRepo(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	belajarRepo := repository.NewBelajarRepository(db)
	gameRepo := repository.NewGameRepository(db)
	laporanBesarRepo := repository.NewLaporanBesarRepository(db)

	authUC := usecase.NewAuthUsecase(userRepo, kelasRepo, jm, cfg, fbScryptConfig(cfg))
	profileUC := usecase.NewProfileUsecase(userRepo, cfg)
	dashboardUC := usecase.NewDashboardUsecase(userRepo, aktivasiRepo, pengerjaanRepo)
	sesiUC := usecase.NewSesiUsecase(sesiRepo, courseRepo)
	soalUC := usecase.NewSoalUsecase(soalRepo, courseRepo, terpilihRepo)
	laporanBesarUC := usecase.NewLaporanBesarUsecase(laporanBesarRepo, userRepo, sbLaporan)
	aktivasiUC := usecase.NewAktivasiUsecase(aktivasiRepo, sesiRepo, courseRepo, kelasRepo, jawabanRepo, pengerjaanRepo, soalUC, aktivasiTxRepo)
	jawabanUC := usecase.NewJawabanUsecase(aktivasiRepo, terpilihRepo, jawabanRepo, pengerjaanRepo, userRepo, courseRepo, laporanBesarUC)
	penilaianUC := usecase.NewPenilaianUsecase(jawabanRepo, pengerjaanRepo, userRepo, penilaianTxRepo)
	konfUC := usecase.NewKonfigurasiUsecase(konfRepo)
	userUC := usecase.NewUserUsecase(userRepo, kelasRepo)
	kelasUC := usecase.NewKelasUsecase(kelasRepo)
	jadwalUC := usecase.NewJadwalUsecase(jadwalRepo)
	pedomanUC := usecase.NewPedomanUsecase(pedomanRepo)
	praktikumUC := usecase.NewPraktikumUsecase(sesiRepo, courseRepo, aktivasiRepo, pengerjaanRepo, userRepo, jadwalRepo, terpilihRepo)
	ampuanUC := usecase.NewAmpuanUsecase(ampuanRepo)
	rekapUC := usecase.NewRekapUsecase(rekapRepo, kelasRepo)
	aiGradingUC := usecase.NewAIGradingUsecase(jawabanRepo, penilaianUC, oc, konfRepo)
	rumpangUC := usecase.NewRumpangUsecase(soalRepo, courseRepo, oc)
	auditLogUC := usecase.NewAuditLogUsecase(auditLogRepo, userRepo)
	belajarUC := usecase.NewBelajarUsecase(belajarRepo)
	gameUC := usecase.NewGameUsecase(gameRepo, belajarRepo)

	return &deliveryhttp.Praktikum{
		Config:      cfg,
		JWT:         jm,
		Users:       userRepo,
		Konfigurasi: konfRepo,
		Handlers: deliveryhttp.PraktikumHandlers{
			Auth:         handler.NewAuthHandler(authUC, profileUC, auditLogUC),
			Dashboard:    handler.NewDashboardHandler(dashboardUC),
			Sesi:         handler.NewSesiHandler(sesiUC, auditLogUC),
			Aktivasi:     handler.NewAktivasiHandler(aktivasiUC, auditLogUC),
			Soal:         handler.NewSoalHandler(soalUC, auditLogUC),
			Jawaban:      handler.NewJawabanHandler(jawabanUC, auditLogUC),
			Penilaian:    handler.NewPenilaianHandler(penilaianUC, auditLogUC),
			Konfigurasi:  handler.NewKonfigurasiHandler(konfUC, cfg.OllamaModel, auditLogUC),
			Profile:      handler.NewProfileHandler(profileUC),
			Jadwal:       handler.NewJadwalHandler(jadwalUC),
			User:         handler.NewUserHandler(userUC, auditLogUC),
			Kelas:        handler.NewKelasHandler(kelasUC),
			Pedoman:      handler.NewPedomanHandler(pedomanUC),
			Praktikum:    handler.NewPraktikumHandler(praktikumUC),
			Upload:       handler.NewUploadHandler(sb),
			Ampuan:       handler.NewAmpuanHandler(ampuanUC, userUC),
			Rekap:        handler.NewRekapHandler(rekapUC),
			AIGrading:    handler.NewAIGradingHandler(aiGradingUC),
			RekapJawaban: handler.NewRekapJawabanHandler(penilaianUC, auditLogUC),
			Cron:         handler.NewCronHandler(jawabanUC, rumpangUC, cfg.CronSecret),
			Run:          handler.NewRunHandler(glotClient),
			AuditLog:     handler.NewAuditLogHandler(auditLogUC),
			Belajar:      handler.NewBelajarHandler(belajarUC),
			Game:         handler.NewGameHandler(gameUC),
			LaporanBesar: handler.NewLaporanBesarHandler(laporanBesarUC),
		},
	}
}
