package usecase

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Konstanta prestigeService.ts.
const (
	prestigeCooldownDays    = 14
	prestigeMaxPerSemester  = 2
	prestigeRequiredLevels  = 6
	prestigeMasterMult      = 1.5
	prestigeTierMaster      = "Master"
	msgPrestigeAccessDenied = "Akses ditolak."
	eventProgressReset      = "progress_reset"
)

func prestigeTierForLevels(levelsDone int) (tier string, mult float64) {
	if levelsDone >= prestigeRequiredLevels {
		return prestigeTierMaster, prestigeMasterMult
	}
	return "Base", 1.0
}

func prestigeBadgeForCount(count int) string {
	if count > 0 {
		return fmt.Sprintf("Prestige-%d", count)
	}
	return "Base"
}

// Semester kalender Jakarta: Ganjil 1 Agu-31 Jan, Genap 1 Feb-31 Jul (semesterStartJakarta).
func semesterStartJakarta(now time.Time) time.Time {
	j := now.UTC().Add(7 * time.Hour)
	y, m := j.Year(), j.Month()
	var start time.Time
	switch {
	case m >= time.February && m <= time.July:
		start = time.Date(y, time.February, 1, 0, 0, 0, 0, time.UTC)
	case m >= time.August:
		start = time.Date(y, time.August, 1, 0, 0, 0, 0, time.UTC)
	default:
		start = time.Date(y-1, time.August, 1, 0, 0, 0, 0, time.UTC)
	}
	return start.Add(-7 * time.Hour)
}

func semesterLabelJakarta(now time.Time) string {
	j := now.UTC().Add(7 * time.Hour)
	y, m := j.Year(), j.Month()
	switch {
	case m >= time.February && m <= time.July:
		return fmt.Sprintf("Genap Feb–Jul %d", y)
	case m >= time.August:
		return fmt.Sprintf("Ganjil Agu %d–Jan %d", y, y+1)
	}
	return fmt.Sprintf("Ganjil Agu %d–Jan %d", y-1, y)
}

type PrestigeUsecase struct {
	users      repository.UserRepository
	curriculum repository.CurriculumRepository
	progress   repository.ProgressRepository
	prestige   repository.PrestigeRepository
	logs       *ActivityLogUsecase
	now        func() time.Time
}

func NewPrestigeUsecase(users repository.UserRepository, curriculum repository.CurriculumRepository, progress repository.ProgressRepository, prestige repository.PrestigeRepository, logs *ActivityLogUsecase, now func() time.Time) *PrestigeUsecase {
	return &PrestigeUsecase{users: users, curriculum: curriculum, progress: progress, prestige: prestige, logs: logs, now: now}
}

// Disalin dari checkPrestigeEligibility dengan saringan RLS token: tabel lessons hanya terbaca staf
// (R-lessons-SELECT-1), progres dan prestige_history orang lain hanya terbaca staf (R-student_progress-SELECT-1,
// R-prestige_history-SELECT-1).
func (uc *PrestigeUsecase) Eligibility(ctx context.Context, p *dto.Principal, nim string) (*dto.PrestigeEligibility, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	nim = cmp.Or(nim, p.NIM)
	tree, err := uc.curriculum.LevelTree(ctx)
	if err != nil {
		return nil, err
	}
	lessonsVisible := IsStaff(p.Role)

	done := map[string]bool{}
	var semesterUsed int64
	state := &repository.PrestigeState{}
	if ownsOrStaff(p, nim) {
		ids, err := uc.progress.CompletedLessonIDs(ctx, nim)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			done[id] = true
		}
		if semesterUsed, err = uc.prestige.CountSince(ctx, nim, semesterStartJakarta(uc.now())); err != nil {
			return nil, err
		}
	}
	if s, err := uc.users.PrestigeState(ctx, nim); err != nil {
		return nil, err
	} else if s != nil {
		state = s
	}

	all := map[string]bool{}
	levelsDone := 0
	for _, lv := range tree {
		var lessons []string
		if lessonsVisible {
			lessons = lv.LessonIDs
		}
		complete := len(lessons) > 0
		for _, id := range lessons {
			all[id] = true
			complete = complete && done[id]
		}
		if complete {
			levelsDone++
		}
	}

	tier, mult := prestigeTierForLevels(levelsDone)
	e := &dto.PrestigeEligibility{
		LevelsDone: levelsDone, TotalLevels: len(tree), LessonsDone: len(done), TotalLessons: len(all),
		Tier: tier, Mult: mult, PrestigeCount: state.Count, LastPrestigeAt: state.LastPrestigeAt, SemesterUsed: semesterUsed,
	}
	switch {
	case levelsDone < prestigeRequiredLevels || len(done) < len(all):
		e.Reason = fmt.Sprintf("Syarat belum terpenuhi: tamat %d/%d level (%d/%d lesson). Reset sukarela hanya untuk yang tamat penuh 6/6.", levelsDone, len(tree), len(done), len(all))
	case state.LastPrestigeAt != nil && uc.now().Sub(*state.LastPrestigeAt).Hours()/24 < prestigeCooldownDays:
		days := uc.now().Sub(*state.LastPrestigeAt).Hours() / 24
		e.Reason = fmt.Sprintf("Cooldown: reset berikutnya bisa dilakukan %d hari lagi (jeda 14 hari antar reset).", int(math.Ceil(prestigeCooldownDays-days)))
	case semesterUsed >= prestigeMaxPerSemester:
		// Label semester berjalan, bukan semester depan, persis kode lama.
		e.Reason = fmt.Sprintf("Kuota semester habis (%d/%d). Reset lagi semester depan (%s).", semesterUsed, prestigeMaxPerSemester, semesterLabelJakarta(uc.now()))
	default:
		e.Eligible = true
		e.Reason = fmt.Sprintf("Layak prestige: tamat penuh 6/6 → Master x1.50 lesson-only. Badge: %s. Papan mingguan tetap base.", prestigeBadgeForCount(state.Count+1))
		e.Tier, e.Mult = prestigeTierMaster, prestigeMasterMult
	}
	return e, nil
}

// Disalin dari doVoluntaryPrestigeReset: guard memakai peran efektif UI, lalu cek kelayakan, lalu RPC apa adanya.
func (uc *PrestigeUsecase) Reset(ctx context.Context, p *dto.Principal, req *dto.PrestigeResetRequest) (*dto.PrestigeResetResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	nim := cmp.Or(req.NIM, p.NIM)
	if nim != p.NIM {
		role, err := uiRole(ctx, uc.users, p)
		if err != nil {
			return nil, err
		}
		if !IsStaff(role) {
			return nil, Forbidden(msgPrestigeAccessDenied)
		}
	}
	elig, err := uc.Eligibility(ctx, p, nim)
	if err != nil {
		return nil, err
	}
	if !elig.Eligible {
		return nil, BadRequest(elig.Reason)
	}
	raw, err := uc.prestige.Reset(ctx, nim)
	if err != nil {
		return nil, MapDBError(err)
	}
	res, err := parseJS([]byte(raw))
	if err != nil {
		return nil, err
	}
	out := &dto.PrestigeResetResponse{
		Tier:       jsToString(or(prop(res, "tier"), prestigeTierMaster)),
		Mult:       jsNumber(or(prop(res, "mult"), prestigeMasterMult)),
		PrestigeNo: jsNumber(or(prop(res, "prestige_no"), 0.0)),
	}
	uc.logs.audit(ctx, p, eventProgressReset, fmt.Sprintf("Prestige sukarela #%s NIM %s: tamat penuh 6/6 → tier %s x%.2f (lesson-only, RPC atomik). Progres diarsip, bukan dihapus.",
		jsNumberString(out.PrestigeNo), nim, out.Tier, out.Mult))
	return out, nil
}
