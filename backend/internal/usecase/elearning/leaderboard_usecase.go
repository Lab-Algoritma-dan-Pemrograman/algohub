package usecase

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Bonus transisi khusus minggu 19-25 Okt 2026 (leaderboardService.ts CARRY_WEEK_LABEL/CARRY_DISCOUNT).
const (
	carryWeekLabel = "2026-10-19"
	carryDiscount  = 0.5
)

type jakartaWeek struct {
	start, end time.Time
	startLabel string
}

// Disalin dari getJakartaWeek: Senin 00:00 WIB sampai Senin berikutnya (eksklusif). Label memakai tanggal kalender
// tanpa normalisasi bulan seperti kode lama (bisa bernilai nol atau negatif di awal bulan); hanya dipakai untuk cek carry.
func jakartaWeekOf(now time.Time) jakartaWeek {
	j := now.UTC().Add(7 * time.Hour)
	sinceMonday := (int(j.Weekday()) + 6) % 7
	y, m, d := j.Date()
	start := time.Date(y, m, d-sinceMonday, 0, 0, 0, 0, time.UTC).Add(-7 * time.Hour)
	return jakartaWeek{start: start, end: start.Add(7 * 24 * time.Hour), startLabel: fmt.Sprintf("%d-%02d-%02d", y, int(m), d-sinceMonday)}
}

type LeaderboardUsecase struct {
	board repository.LeaderboardRepository
	now   func() time.Time
}

func NewLeaderboardUsecase(board repository.LeaderboardRepository, now func() time.Time) *LeaderboardUsecase {
	return &LeaderboardUsecase{board: board, now: now}
}

func entryOf(nim string, nama, kelas, jurusan, shift *string, xp, level, streak *int) dto.LeaderboardEntry {
	return dto.LeaderboardEntry{
		NIM: nim, Nama: nama, Kelas: kelas, Jurusan: jurusan, Shift: shift,
		XP:     intValue(xp),
		Level:  cmp.Or(intValue(level), 1),
		Streak: intValue(streak),
		Role:   RolePraktikan, AssessmentAccess: map[string]any{},
	}
}

func intValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// Disalin dari getLeaderboard: role kanonik mahasiswa, urut xp desc.
func (uc *LeaderboardUsecase) List(ctx context.Context, p *dto.Principal, f repository.LeaderboardFilter, page, limit int) (*dto.Page[dto.LeaderboardEntry], error) {
	return listPage(p, page, limit, func() ([]dto.LeaderboardEntry, int64, error) {
		rows, total, err := uc.board.Leaderboard(ctx, f, page, limit)
		out := make([]dto.LeaderboardEntry, len(rows))
		for i, u := range rows {
			out[i] = entryOf(u.NIM, u.Nama, u.Kelas, u.Jurusan, u.Shift, u.XP, u.Level, u.Streak)
		}
		return out, total, err
	})
}

// Disalin dari getUserRank: xp 0 tidak punya peringkat.
func (uc *LeaderboardUsecase) Rank(ctx context.Context, p *dto.Principal, xp float64, f repository.LeaderboardFilter) (*dto.RankResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	if xp == 0 {
		return &dto.RankResponse{}, nil
	}
	n, err := uc.board.CountAbove(ctx, xp, f)
	if err != nil {
		return nil, err
	}
	rank := n + 1
	return &dto.RankResponse{Rank: &rank}, nil
}

// Disalin dari getLeaderboardFilters: nilai kosong dibuang, diurutkan seperti Array.prototype.sort.
func (uc *LeaderboardUsecase) Filters(ctx context.Context, p *dto.Principal) (*dto.LeaderboardFilters, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	rows, err := uc.board.Filters(ctx)
	if err != nil {
		return nil, err
	}
	out := &dto.LeaderboardFilters{Kelas: []string{}, Jurusan: []string{}, Shift: []string{}}
	lists := map[string]*[]string{"kelas": &out.Kelas, "jurusan": &out.Jurusan, "shift": &out.Shift}
	for _, r := range rows {
		if l, ok := lists[r.Kind]; ok && r.Value != "" && !slices.Contains(*l, r.Value) {
			*l = append(*l, r.Value)
		}
	}
	for _, l := range lists {
		slices.Sort(*l)
	}
	return out, nil
}

// Disalin dari getWeeklyLeaderboard. Penampil non-staf hanya melihat agregat miliknya dengan reward 60 (keputusan 1).
func (uc *LeaderboardUsecase) Weekly(ctx context.Context, p *dto.Principal, f repository.LeaderboardFilter, page, limit int) (*dto.Page[dto.WeeklyEntry], error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	var viewer *string
	if !IsStaff(p.Role) {
		viewer = &p.AuthNIM
	}
	week := jakartaWeekOf(uc.now())
	rows, err := uc.board.Weekly(ctx, f, week.start, week.end, viewer)
	if err != nil {
		return nil, err
	}
	entries := make([]dto.WeeklyEntry, len(rows))
	for i, r := range rows {
		total := int(r.LessonXP + r.BughuntXP)
		entries[i] = dto.WeeklyEntry{
			LeaderboardEntry: entryOf(r.NIM, r.Nama, r.Kelas, r.Jurusan, r.Shift, &total, r.Level, r.Streak),
			WeeklyXP:         total,
			LessonXP:         int(r.LessonXP),
			BughuntXP:        int(r.BughuntXP),
			LessonsCompleted: int(r.LessonsCompleted),
		}
	}
	if week.startLabel == carryWeekLabel {
		if err := uc.applyCarry(ctx, f, week, viewer, entries); err != nil {
			return nil, err
		}
	}
	sortWeekly(entries)
	return pageOf(paginate(entries, page, limit), int64(len(entries)), page, limit), nil
}

// Semua entry dengan XP minggu lalu > 0 mendapat floor(50%) sebagai bonus di atas XP minggu ini.
func (uc *LeaderboardUsecase) applyCarry(ctx context.Context, f repository.LeaderboardFilter, week jakartaWeek, viewer *string, entries []dto.WeeklyEntry) error {
	prev, err := uc.board.Weekly(ctx, f, week.start.Add(-7*24*time.Hour), week.start, viewer)
	if err != nil {
		return err
	}
	prevTotal := map[string]int64{}
	for _, r := range prev {
		prevTotal[r.NIM] = r.LessonXP + r.BughuntXP
	}
	for i := range entries {
		e := &entries[i]
		if t := prevTotal[e.NIM]; t > 0 {
			bonus := int(math.Floor(float64(t) * carryDiscount))
			e.WeeklyXP += bonus
			e.XP = e.WeeklyXP
			e.IsCarry = true
			e.CarryXP = bonus
		}
	}
	return nil
}

// Urutan `b.weeklyXp - a.weeklyXp || a.nama.localeCompare(b.nama)`; collation root Unicode mendekati localeCompare peramban.
func sortWeekly(entries []dto.WeeklyEntry) {
	col := collate.New(language.Und)
	slices.SortStableFunc(entries, func(a, b dto.WeeklyEntry) int {
		if c := cmp.Compare(b.WeeklyXP, a.WeeklyXP); c != 0 {
			return c
		}
		return col.CompareString(strValue(a.Nama), strValue(b.Nama))
	})
}

func paginate[T any](items []T, page, limit int) []T {
	from := min((page-1)*limit, len(items))
	return items[from:min(from+limit, len(items))]
}
