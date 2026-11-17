package usecase

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"math"
	"slices"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Salinan src/data/achievements.json: cadangan checkAndUnlockAchievements dan getAchievements bila tabel kosong atau gagal dibaca.
//
//go:embed data/achievements.json
var defaultAchievementsJSON []byte

var defaultAchievements = mustAchievements(defaultAchievementsJSON)

func mustAchievements(raw []byte) []dto.Achievement {
	var out []dto.Achievement
	if err := json.Unmarshal(raw, &out); err != nil {
		panic("achievements.json rusak: " + err.Error())
	}
	return out
}

type AchievementUsecase struct {
	achievements repository.AchievementRepository
	users        repository.UserRepository
	progress     repository.ProgressRepository
	now          func() time.Time
}

func NewAchievementUsecase(achievements repository.AchievementRepository, users repository.UserRepository, progress repository.ProgressRepository, now func() time.Time) *AchievementUsecase {
	return &AchievementUsecase{achievements: achievements, users: users, progress: progress, now: now}
}

// requirement_value disimpan teks; isNaN(Number(v)) ? v : Number(v).
func formatAchievement(a *entity.Achievement) dto.Achievement {
	value := nullable(a.RequirementValue)
	if n := jsNumber(value); !math.IsNaN(n) {
		value = n
	}
	return dto.Achievement{ID: a.ID, Title: a.Title, Description: a.Description, Icon: a.Icon, RequirementType: a.RequirementType, RequirementValue: value}
}

// getAchievements: tabel kosong atau error berarti cadangan achievements.json.
func (uc *AchievementUsecase) catalog(ctx context.Context, requirementType string) []dto.Achievement {
	rows, err := uc.achievements.List(ctx, requirementType)
	if err == nil && len(rows) > 0 {
		out := make([]dto.Achievement, len(rows))
		for i := range rows {
			out[i] = formatAchievement(&rows[i])
		}
		return out
	}
	if err != nil {
		log.Printf("elearning: baca achievements: %v", err)
	}
	if requirementType == "" {
		return defaultAchievements
	}
	var out []dto.Achievement
	for _, a := range defaultAchievements {
		if a.RequirementType != nil && *a.RequirementType == requirementType {
			out = append(out, a)
		}
	}
	return out
}

func (uc *AchievementUsecase) List(ctx context.Context, p *dto.Principal, requirementType string) ([]dto.Achievement, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	out := uc.catalog(ctx, requirementType)
	if out == nil {
		out = []dto.Achievement{}
	}
	return out, nil
}

func (uc *AchievementUsecase) unlockedIDs(ctx context.Context, p *dto.Principal, nim string) ([]string, error) {
	if !ownsOrStaff(p, nim) {
		return []string{}, nil
	}
	return uc.achievements.UnlockedIDs(ctx, nim)
}

func (uc *AchievementUsecase) Mine(ctx context.Context, p *dto.Principal) ([]string, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	return uc.unlockedIDs(ctx, p, p.NIM)
}

// WITH CHECK unlocked_achievements hanya nim = auth_nim() untuk semua peran, termasuk admin/kordas.
func (uc *AchievementUsecase) insert(ctx context.Context, p *dto.Principal, nim, achievementID string) error {
	if nim != p.AuthNIM {
		return RLSInsert("unlocked_achievements")
	}
	return uc.achievements.Insert(ctx, nim, achievementID, uc.now().UTC())
}

type achievementUser struct {
	nim    string
	xp     float64
	streak int
}

// Disalin dari checkAndUnlockAchievements; setiap error ditelan dan menghasilkan daftar kosong atau dilewati, seperti kode lama.
func (uc *AchievementUsecase) unlock(ctx context.Context, p *dto.Principal, u achievementUser, stats *dto.AchievementCheckRequest) []dto.Achievement {
	newly := []dto.Achievement{}
	if u.nim == "" {
		return newly
	}
	all := uc.catalog(ctx, "")
	ids, err := uc.unlockedIDs(ctx, p, u.nim)
	if err != nil {
		log.Printf("elearning: baca unlocked_achievements: %v", err)
	}
	var lessonCount float64
	if stats.LessonCount != nil {
		lessonCount = *stats.LessonCount
	} else if ownsOrStaff(p, u.nim) {
		n, err := uc.progress.CountCompleted(ctx, u.nim)
		if err != nil {
			log.Printf("elearning: hitung progres: %v", err)
		}
		lessonCount = float64(n)
	}
	xp := u.xp
	if stats.XP != nil {
		xp = *stats.XP
	}
	for i := range all {
		a := &all[i]
		if slices.Contains(ids, a.ID) || !achievementMet(a, xp, float64(u.streak), lessonCount, stats) {
			continue
		}
		if err := uc.insert(ctx, p, u.nim, a.ID); err != nil {
			log.Printf("elearning: unlock %s: %v", a.ID, err)
			continue
		}
		newly = append(newly, *a)
	}
	return newly
}

func achievementMet(a *dto.Achievement, xp, streak, lessonCount float64, stats *dto.AchievementCheckRequest) bool {
	if a.RequirementType == nil {
		return false
	}
	want := jsNumber(a.RequirementValue)
	switch *a.RequirementType {
	case "xp":
		return xp >= want
	case "streak":
		return streak >= want
	case "level_completed":
		s, ok := a.RequirementValue.(string)
		return ok && slices.Contains(stats.CompletedLevelIDs, s)
	case "lesson_count":
		return lessonCount >= want
	case "game_score":
		perfect := deref(stats.PerfectGames)
		if a.ID == "perfect-hunt" {
			return perfect >= want
		}
		return math.Max(deref(stats.GamesPlayed), perfect) >= want
	}
	return false
}

func deref(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

// user dulu diambil dari store browser; kini dari users di DB.
func (uc *AchievementUsecase) Check(ctx context.Context, p *dto.Principal, stats *dto.AchievementCheckRequest) ([]dto.Achievement, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	u := achievementUser{nim: p.NIM}
	profile, err := uc.users.GetProfile(ctx, p.NIM)
	if err != nil {
		log.Printf("elearning: baca profil: %v", err)
		return []dto.Achievement{}, nil
	}
	if profile != nil {
		u.xp, u.streak = float64(profile.XP), profile.Streak
	}
	return uc.unlock(ctx, p, u, stats), nil
}

func (uc *AchievementUsecase) CheckXP(ctx context.Context, p *dto.Principal, req *dto.CheckXPRequest) (*dto.CheckXPResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	return &dto.CheckXPResponse{Achievement: uc.checkXP(ctx, p, p.NIM, req.XP)}, nil
}

// Disalin dari checkXpAchievements: tanpa cadangan JSON, urut ambang naik, kembalikan yang pertama berhasil dicatat.
func (uc *AchievementUsecase) checkXP(ctx context.Context, p *dto.Principal, nim string, xp float64) *dto.Achievement {
	rows, err := uc.achievements.List(ctx, "xp")
	if err != nil {
		return nil
	}
	ids, err := uc.unlockedIDs(ctx, p, nim)
	if err != nil {
		return nil
	}
	list := make([]dto.Achievement, len(rows))
	for i := range rows {
		list[i] = formatAchievement(&rows[i])
		list[i].RequirementValue = jsNumber(nullable(rows[i].RequirementValue))
	}
	// Pembanding a - b ala Array.prototype.sort: NaN dianggap sama.
	slices.SortStableFunc(list, func(a, b dto.Achievement) int {
		d := a.RequirementValue.(float64) - b.RequirementValue.(float64)
		switch {
		case d < 0:
			return -1
		case d > 0:
			return 1
		}
		return 0
	})
	for i := range list {
		if slices.Contains(ids, list[i].ID) {
			continue
		}
		if xp >= list[i].RequirementValue.(float64) && uc.insert(ctx, p, nim, list[i].ID) == nil {
			return &list[i]
		}
	}
	return nil
}
