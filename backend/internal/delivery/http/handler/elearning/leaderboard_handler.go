package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Bawaan limitCount = 10 di getLeaderboard/getWeeklyLeaderboard.
const defaultLeaderboardLimit = 10

type LeaderboardHandler struct {
	uc *usecase.LeaderboardUsecase
}

func NewLeaderboardHandler(uc *usecase.LeaderboardUsecase) *LeaderboardHandler {
	return &LeaderboardHandler{uc: uc}
}

func leaderboardFilter(c *gin.Context) repository.LeaderboardFilter {
	return repository.LeaderboardFilter{Kelas: c.Query("kelas"), Jurusan: c.Query("jurusan"), Shift: c.Query("shift")}
}

// limit <= 0 dulu berarti semua baris; kini menjadi halaman berukuran maksimum (keputusan 14).
func leaderboardPage(c *gin.Context) (page, limit int) {
	page, _ = pageParams(c)
	v, err := strconv.Atoi(c.Query("limit"))
	switch {
	case err != nil:
		return page, defaultLeaderboardLimit
	case v <= 0:
		return page, maxPageLimit
	}
	return page, min(v, maxPageLimit)
}

func (h *LeaderboardHandler) List(c *gin.Context) {
	f := leaderboardFilter(c)
	page, limit := leaderboardPage(c)
	serve(c, "Leaderboard", func(ctx context.Context, p *dto.Principal) (*dto.Page[dto.LeaderboardEntry], error) {
		return h.uc.List(ctx, p, f, page, limit)
	})
}

func (h *LeaderboardHandler) Weekly(c *gin.Context) {
	f := leaderboardFilter(c)
	page, limit := leaderboardPage(c)
	serve(c, "Leaderboard mingguan", func(ctx context.Context, p *dto.Principal) (*dto.Page[dto.WeeklyEntry], error) {
		return h.uc.Weekly(ctx, p, f, page, limit)
	})
}

// xp tidak valid dulu membuat query PostgREST gagal dan getUserRank mengembalikan null.
func (h *LeaderboardHandler) Rank(c *gin.Context) {
	f := leaderboardFilter(c)
	xp, err := strconv.ParseFloat(c.Query("xp"), 64)
	if err != nil {
		xp = 0
	}
	serve(c, "Peringkat", func(ctx context.Context, p *dto.Principal) (*dto.RankResponse, error) {
		return h.uc.Rank(ctx, p, xp, f)
	})
}

func (h *LeaderboardHandler) Filters(c *gin.Context) {
	serve(c, "Filter leaderboard", h.uc.Filters)
}
