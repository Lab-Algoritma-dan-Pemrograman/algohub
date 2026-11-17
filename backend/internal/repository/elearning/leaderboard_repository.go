package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Filter kelas/jurusan/shift di leaderboardService.ts hanya dipasang bila nilainya tidak kosong.
type LeaderboardFilter struct {
	Kelas   string
	Jurusan string
	Shift   string
}

// Kolom persis select leaderboardService.ts:307.
type LeaderboardUser struct {
	NIM     string  `gorm:"column:nim"`
	Nama    *string `gorm:"column:nama"`
	Kelas   *string `gorm:"column:kelas"`
	Jurusan *string `gorm:"column:jurusan"`
	Shift   *string `gorm:"column:shift"`
	XP      *int    `gorm:"column:xp"`
	Level   *int    `gorm:"column:level"`
	Streak  *int    `gorm:"column:streak"`
}

func (LeaderboardUser) TableName() string { return "users" }

type FilterValue struct {
	Kind  string `gorm:"column:kind"`
	Value string `gorm:"column:value"`
}

type WeeklyRow struct {
	NIM              string  `gorm:"column:nim"`
	Nama             *string `gorm:"column:nama"`
	Kelas            *string `gorm:"column:kelas"`
	Jurusan          *string `gorm:"column:jurusan"`
	Shift            *string `gorm:"column:shift"`
	Level            *int    `gorm:"column:level"`
	Streak           *int    `gorm:"column:streak"`
	LessonXP         int64   `gorm:"column:lesson_xp"`
	LessonsCompleted int64   `gorm:"column:lessons_completed"`
	BughuntXP        int64   `gorm:"column:bughunt_xp"`
}

type LeaderboardRepository interface {
	Leaderboard(ctx context.Context, f LeaderboardFilter, page, limit int) ([]LeaderboardUser, int64, error)
	CountAbove(ctx context.Context, xp float64, f LeaderboardFilter) (int64, error)
	Filters(ctx context.Context) ([]FilterValue, error)
	Weekly(ctx context.Context, f LeaderboardFilter, start, end time.Time, viewer *string) ([]WeeklyRow, error)
}

type leaderboardRepository struct{ db *gorm.DB }

func NewLeaderboardRepository(db *gorm.DB) LeaderboardRepository {
	return &leaderboardRepository{db: db}
}

func (f LeaderboardFilter) apply(db *gorm.DB) *gorm.DB {
	db = db.Where("role = ?", "mahasiswa")
	if f.Kelas != "" {
		db = db.Where("kelas = ?", f.Kelas)
	}
	if f.Jurusan != "" {
		db = db.Where("jurusan = ?", f.Jurusan)
	}
	if f.Shift != "" {
		db = db.Where("shift = ?", f.Shift)
	}
	return db
}

// order('xp', desc) PostgREST memakai NULLS FIRST bawaan Postgres; nim menjadi pemecah seri supaya halaman stabil.
func (r *leaderboardRepository) Leaderboard(ctx context.Context, f LeaderboardFilter, page, limit int) ([]LeaderboardUser, int64, error) {
	return findPage[LeaderboardUser](ctx, f.apply, r.db, "xp DESC, nim", page, limit)
}

func (r *leaderboardRepository) CountAbove(ctx context.Context, xp float64, f LeaderboardFilter) (int64, error) {
	var n int64
	err := f.apply(r.db.WithContext(ctx).Model(&LeaderboardUser{})).Where("xp > ?", xp).Count(&n).Error
	return n, err
}

// Satu query DISTINCT menggantikan pembacaan seluruh baris mahasiswa di getLeaderboardFilters.
func (r *leaderboardRepository) Filters(ctx context.Context) ([]FilterValue, error) {
	var rows []FilterValue
	err := r.db.WithContext(ctx).Raw(`SELECT 'kelas' AS kind, kelas::text AS value FROM users WHERE role = 'mahasiswa' AND kelas IS NOT NULL
UNION SELECT 'jurusan', jurusan::text FROM users WHERE role = 'mahasiswa' AND jurusan IS NOT NULL
UNION SELECT 'shift', shift::text FROM users WHERE role = 'mahasiswa' AND shift IS NOT NULL`).Scan(&rows).Error
	return rows, err
}

// Satu SQL agregasi menggantikan potongan 100 NIM x dua query di getWeeklyLeaderboard. viewer != nil meniru RLS untuk
// penampil non-staf: hanya baris student_progress/game_history miliknya yang terbaca dan embed lessons:lesson_id(xp_reward)
// bernilai null sehingga reward jatuh ke `?? 60` (R-student_progress-SELECT-1, R-game_history-SELECT-1, R-lessons-SELECT-2).
func (r *leaderboardRepository) Weekly(ctx context.Context, f LeaderboardFilter, start, end time.Time, viewer *string) ([]WeeklyRow, error) {
	args := map[string]any{"start": start, "end": end}
	cand := []string{"role = 'mahasiswa'"}
	for _, c := range []struct{ col, val string }{{"kelas", f.Kelas}, {"jurusan", f.Jurusan}, {"shift", f.Shift}} {
		if c.val != "" {
			cand = append(cand, c.col+" = @"+c.col)
			args[c.col] = c.val
		}
	}
	reward, lessonJoin, progOwner, gameOwner := "COALESCE(l.xp_reward, 60)", "LEFT JOIN lessons l ON l.id = sp.lesson_id", "", ""
	if viewer != nil {
		reward, lessonJoin = "60", ""
		progOwner, gameOwner = " AND sp.nim = @viewer", " AND gh.nim = @viewer"
		args["viewer"] = *viewer
	}
	sql := `WITH cand AS (
  SELECT nim, nama, kelas, jurusan, shift::text AS shift, level, streak FROM users WHERE ` + strings.Join(cand, " AND ") + `
), prog AS (
  SELECT sp.nim, COUNT(*) AS lessons_completed, SUM(` + reward + `) AS lesson_xp
  FROM student_progress sp ` + lessonJoin + `
  WHERE sp.completed AND sp.completed_at >= @start AND sp.completed_at < @end AND sp.nim IN (SELECT nim FROM cand)` + progOwner + `
  GROUP BY sp.nim
), games AS (
  SELECT gh.nim, SUM(COALESCE(gh.xp_earned, 0)) AS bughunt_xp
  FROM game_history gh
  WHERE gh.played_at >= @start AND gh.played_at < @end AND gh.nim IN (SELECT nim FROM cand)` + gameOwner + `
  GROUP BY gh.nim
)
SELECT c.nim, c.nama, c.kelas, c.jurusan, c.shift, c.level, c.streak,
  COALESCE(p.lesson_xp, 0) AS lesson_xp, COALESCE(p.lessons_completed, 0) AS lessons_completed, COALESCE(g.bughunt_xp, 0) AS bughunt_xp
FROM cand c
LEFT JOIN prog p ON p.nim = c.nim
LEFT JOIN games g ON g.nim = c.nim
ORDER BY c.nim`
	var rows []WeeklyRow
	err := r.db.WithContext(ctx).Raw(sql, args).Scan(&rows).Error
	return rows, err
}
