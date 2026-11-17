package usecase

import (
	"cmp"
	"context"
	"log"
	"math"
	"slices"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Hari kalender streak dulu dihitung di zona waktu browser mahasiswa (WIB), kini di server (keputusan 13).
var wib = time.FixedZone("WIB", 7*3600)

// Batas PRESTIGE_MULT_CAP di prestigeService.ts.
const prestigeMultCap = 1.5

// Disalin dari calculateLevel di progressService.ts; sama dengan public.calculate_level.
func calculateLevel(xp int) int {
	if xp <= 0 {
		return 1
	}
	return int(math.Floor(math.Sqrt(float64(xp)/50))) + 1
}

// Disalin dari applyPrestigeMult di prestigeService.ts; `mult || 1` membuat 0 dan NaN menjadi 1.
func applyPrestigeMult(base, mult float64) int {
	if mult == 0 || math.IsNaN(mult) {
		mult = 1
	}
	return int(math.Floor(base * math.Min(prestigeMultCap, math.Max(1, mult))))
}

// Disalin dari calculateStreak di streakService.ts.
func calculateStreak(lastActive *time.Time, current int, now time.Time) (streak int, changed bool) {
	safe := max(0, current)
	if lastActive == nil {
		return 1, safe != 1
	}
	diff := int(math.Round(midnightWIB(now).Sub(midnightWIB(*lastActive)).Hours() / 24))
	switch {
	case diff == 1:
		streak = safe + 1
	case diff > 1:
		streak = 1
	default:
		streak = max(safe, 1)
	}
	return streak, streak != safe
}

func midnightWIB(t time.Time) time.Time {
	y, m, d := t.In(wib).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, wib)
}

type ProgressUsecase struct {
	users        repository.UserRepository
	progress     repository.ProgressRepository
	curriculum   repository.CurriculumRepository
	achievements *AchievementUsecase
	guard        userGuard
	now          func() time.Time
}

func NewProgressUsecase(users repository.UserRepository, progress repository.ProgressRepository, curriculum repository.CurriculumRepository, achievements *AchievementUsecase, now func() time.Time) *ProgressUsecase {
	return &ProgressUsecase{users: users, progress: progress, curriculum: curriculum, achievements: achievements, guard: userGuard{users: users}, now: now}
}

func (uc *ProgressUsecase) MyLessons(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[string], error) {
	return listPage(p, page, limit, func() ([]string, int64, error) {
		if !ownsOrStaff(p, p.NIM) {
			return nil, 0, nil
		}
		rows, total, err := uc.progress.ListByNIM(ctx, p.NIM, page, limit)
		ids := make([]string, len(rows))
		for i := range rows {
			ids[i] = rows[i].LessonID
		}
		return ids, total, err
	})
}

func (uc *ProgressUsecase) Recent(ctx context.Context, p *dto.Principal, limit int) ([]entity.StudentProgress, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	if !ownsOrStaff(p, p.NIM) {
		return []entity.StudentProgress{}, nil
	}
	rows, err := uc.progress.Recent(ctx, p.NIM, limit)
	if rows == nil {
		rows = []entity.StudentProgress{}
	}
	return rows, err
}

// Pengganti store user di browser: nilai diambil dari DB, profil tidak ada berarti nilai awal.
func (uc *ProgressUsecase) profile(ctx context.Context, nim string) (*entity.User, error) {
	u, err := uc.users.GetProfile(ctx, nim)
	if err != nil {
		return nil, err
	}
	if u == nil {
		u = &entity.User{NIM: nim, PrestigeMult: 1}
	}
	return u, nil
}

// Disalin dari completeLesson (progressService.ts:43): progres dulu, baru users, supaya trigger XP sudah melihat baris baru.
func (uc *ProgressUsecase) Complete(ctx context.Context, p *dto.Principal, lessonID string, req *dto.CompleteLessonRequest) (*dto.CompleteLessonResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	nim := p.NIM
	u, err := uc.profile(ctx, nim)
	if err != nil {
		return nil, err
	}
	if ownsOrStaff(p, nim) {
		done, err := uc.progress.Exists(ctx, nim, lessonID)
		if err != nil {
			return nil, err
		}
		if done {
			return &dto.CompleteLessonResponse{XP: u.XP, Level: u.Level, Streak: u.Streak, LastActive: u.LastActive, AlreadyCompleted: true, NewlyUnlocked: []dto.Achievement{}}, nil
		}
	}

	now := uc.now().UTC()
	streak, _ := calculateStreak(u.LastActive, u.Streak, now)
	newXP := u.XP + applyPrestigeMult(req.XPReward, u.PrestigeMult)
	newLevel := calculateLevel(newXP)

	if !ownsOrStaff(p, nim) {
		return nil, RLSInsert("student_progress")
	}
	if err := uc.progress.Insert(ctx, nim, lessonID, now); err != nil {
		return nil, MapDBError(err)
	}
	cols := map[string]any{"xp": newXP, "level": newLevel, "streak": streak, "last_active": now}
	if _, err := uc.guard.update(ctx, p, nim, cols); err != nil {
		return nil, err
	}

	return &dto.CompleteLessonResponse{
		XP:            newXP,
		Level:         newLevel,
		Streak:        streak,
		LastActive:    &now,
		LeveledUp:     newLevel > cmp.Or(u.Level, 1),
		NewlyUnlocked: uc.lessonAchievements(ctx, p, lessonID, newXP, streak),
	}, nil
}

// Error tahap pencapaian ditelan seperti try/catch completeLesson; progres dan XP sudah tersimpan.
func (uc *ProgressUsecase) lessonAchievements(ctx context.Context, p *dto.Principal, lessonID string, xp, streak int) []dto.Achievement {
	completed, err := uc.progress.CompletedLessonIDs(ctx, p.NIM)
	if err != nil {
		log.Printf("elearning: baca progres: %v", err)
		return []dto.Achievement{}
	}
	if !slices.Contains(completed, lessonID) {
		completed = append(completed, lessonID)
	}
	tree, err := uc.curriculum.LevelTree(ctx)
	if err != nil {
		log.Printf("elearning: baca kurikulum: %v", err)
		return []dto.Achievement{}
	}
	titles := completedLevelTitles(lessonID, tree, completed)
	var levelIDs []string
	for _, l := range tree {
		if slices.Contains(titles, l.Title) {
			levelIDs = append(levelIDs, l.ID)
		}
	}
	count := float64(len(completed))
	xpf := float64(xp)
	stats := &dto.AchievementCheckRequest{XP: &xpf, LessonCount: &count, CompletedLevelIDs: levelIDs}
	return uc.achievements.unlock(ctx, p, achievementUser{nim: p.NIM, xp: xpf, streak: streak}, stats)
}

// Bagian completedLevels dari getOverallProgress (centralApiService.ts:20); lessonID dihitung selesai.
func completedLevelTitles(lessonID string, tree []repository.LevelLessons, completed []string) []string {
	var titles []string
	for _, l := range tree {
		done := 0
		for _, id := range l.LessonIDs {
			if id == lessonID || slices.Contains(completed, id) {
				done++
			}
		}
		if len(l.LessonIDs) > 0 && done >= len(l.LessonIDs) {
			titles = append(titles, l.Title)
		}
	}
	return titles
}

// Disalin dari grantXp (progressService.ts:157).
func (uc *ProgressUsecase) GrantXP(ctx context.Context, p *dto.Principal, req *dto.GrantXPRequest) (*dto.XPResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	u, err := uc.profile(ctx, p.NIM)
	if err != nil {
		return nil, err
	}
	newXP := u.XP + req.Amount
	newLevel := calculateLevel(newXP)
	cols := map[string]any{"xp": newXP, "level": newLevel, "last_active": uc.now().UTC()}
	if _, err := uc.guard.update(ctx, p, p.NIM, cols); err != nil {
		return nil, err
	}
	return &dto.XPResponse{
		XP:          newXP,
		Level:       newLevel,
		LeveledUp:   newLevel > cmp.Or(u.Level, 1),
		Achievement: uc.achievements.checkXP(ctx, p, p.NIM, float64(newXP)),
	}, nil
}

func (uc *ProgressUsecase) StudyTime(ctx context.Context, p *dto.Principal, req *dto.StudyTimeRequest) (*dto.StudyTimeResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	total, err := uc.guard.incrementStudyTime(ctx, p, p.NIM, req.Seconds)
	if err != nil {
		return nil, err
	}
	return &dto.StudyTimeResponse{StudyTime: total}, nil
}

// Heartbeat Layout.tsx:208-220: last_active selalu ditulis, streak hanya bila berubah.
func (uc *ProgressUsecase) Streak(ctx context.Context, p *dto.Principal) (*dto.StreakResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	u, err := uc.profile(ctx, p.NIM)
	if err != nil {
		return nil, err
	}
	now := uc.now().UTC()
	streak, changed := calculateStreak(u.LastActive, u.Streak, now)
	cols := map[string]any{"last_active": now}
	if changed {
		cols["streak"] = streak
	} else {
		streak = u.Streak
	}
	if _, err := uc.guard.update(ctx, p, p.NIM, cols); err != nil {
		return nil, err
	}
	return &dto.StreakResponse{Streak: streak, LastActive: now}, nil
}
