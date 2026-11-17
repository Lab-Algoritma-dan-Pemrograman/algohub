package repository

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

// Dikembalikan saat trigger guard_prestige_columns (migration-prestige.sql:115) akan menolak UPDATE:
// check_user_xp_level meng-clamp prestige_mult, jadi baris dengan nilai lama di luar 1.00-1.50 selalu berubah.
var ErrPrestigeGuard = errors.New("prestige_mult di luar 1.00-1.50")

// Kolom persis select SupabaseProvider.tsx:92.
const profileColumns = "id,nim,nama,kelas,jurusan,email,role,elearning_role,xp,level,streak,study_time,last_active,created_at,assessment_access,level_access_overrides,tersembunyi,is_aktif,kelas_id,shift,divisi,kelompok,gelombang,prestige_tier,prestige_mult,prestige_count"

// Kolom persis select AdminDashboard.tsx:1221.
type AdminUser struct {
	ID                   *string        `gorm:"column:id" json:"id"`
	NIM                  string         `gorm:"column:nim" json:"nim"`
	Nama                 *string        `gorm:"column:nama" json:"nama"`
	Kelas                *string        `gorm:"column:kelas" json:"kelas"`
	Jurusan              *string        `gorm:"column:jurusan" json:"jurusan"`
	Email                *string        `gorm:"column:email" json:"email"`
	Role                 *string        `gorm:"column:role" json:"role"`
	ElearningRole        *string        `gorm:"column:elearning_role" json:"elearning_role"`
	XP                   *int           `gorm:"column:xp" json:"xp"`
	Level                *int           `gorm:"column:level" json:"level"`
	Streak               *int           `gorm:"column:streak" json:"streak"`
	StudyTime            *int           `gorm:"column:study_time" json:"study_time"`
	LastActive           *time.Time     `gorm:"column:last_active" json:"last_active"`
	CreatedAt            *time.Time     `gorm:"column:created_at" json:"created_at"`
	AssessmentAccess     datatypes.JSON `gorm:"column:assessment_access" json:"assessment_access"`
	LevelAccessOverrides datatypes.JSON `gorm:"column:level_access_overrides" json:"level_access_overrides"`
}

func (AdminUser) TableName() string { return "users" }

// Kolom persis select StudentMonitoring.tsx:82.
type MonitoringStudent struct {
	NIM                  string         `gorm:"column:nim" json:"nim"`
	Nama                 *string        `gorm:"column:nama" json:"nama"`
	Kelas                *string        `gorm:"column:kelas" json:"kelas"`
	Jurusan              *string        `gorm:"column:jurusan" json:"jurusan"`
	Email                *string        `gorm:"column:email" json:"email"`
	Role                 *string        `gorm:"column:role" json:"role"`
	XP                   *int           `gorm:"column:xp" json:"xp"`
	Level                *int           `gorm:"column:level" json:"level"`
	Streak               *int           `gorm:"column:streak" json:"streak"`
	StudyTime            *int           `gorm:"column:study_time" json:"study_time"`
	LastActive           *time.Time     `gorm:"column:last_active" json:"last_active"`
	AssessmentAccess     datatypes.JSON `gorm:"column:assessment_access" json:"assessment_access"`
	LevelAccessOverrides datatypes.JSON `gorm:"column:level_access_overrides" json:"level_access_overrides"`
}

func (MonitoringStudent) TableName() string { return "users" }

type XPLevel struct {
	XP    int
	Level int
}

type PrestigeState struct {
	Count          int        `gorm:"column:prestige_count"`
	LastPrestigeAt *time.Time `gorm:"column:last_prestige_at"`
}

type UserRepository interface {
	RoleByNIM(ctx context.Context, nim string) (string, error)
	RoleByEmail(ctx context.Context, email string) (string, error)
	NamaByNIM(ctx context.Context, nim string) (string, error)
	GetProfile(ctx context.Context, nim string) (*entity.User, error)
	Insert(ctx context.Context, cols map[string]any, recomputeXP bool) error
	UpdateUser(ctx context.Context, nim string, cols map[string]any, recomputeXP bool) (int64, error)
	IncrementStudyTime(ctx context.Context, nim string, seconds int64, recomputeXP bool) (*int64, error)
	ListAdmin(ctx context.Context, page, limit int, q string) ([]AdminUser, int64, error)
	Delete(ctx context.Context, nim string) (int64, error)
	XPLevel(ctx context.Context, nim string) (*XPLevel, error)
	ListStudents(ctx context.Context, page, limit int) ([]MonitoringStudent, int64, error)
	PrestigeState(ctx context.Context, nim string) (*PrestigeState, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) RoleByNIM(ctx context.Context, nim string) (string, error) {
	return r.single(ctx, "COALESCE(role, '')", "nim = ?", nim)
}

func (r *userRepository) RoleByEmail(ctx context.Context, email string) (string, error) {
	return r.single(ctx, "COALESCE(role, '')", "email = ?", email)
}

func (r *userRepository) NamaByNIM(ctx context.Context, nim string) (string, error) {
	return r.single(ctx, "COALESCE(nama, '')", "nim = ?", nim)
}

// Meniru .single()/.maybeSingle() PostgREST: nol atau lebih dari satu baris sama-sama dianggap tidak ditemukan.
func (r *userRepository) single(ctx context.Context, column, where string, arg any) (string, error) {
	var vals []string
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where(where, arg).Limit(2).Pluck(column, &vals).Error; err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", gorm.ErrRecordNotFound
	}
	return vals[0], nil
}

// Hasil nil berarti tidak tepat satu baris; SupabaseProvider memperlakukan error .maybeSingle() sama dengan profil kosong.
func (r *userRepository) GetProfile(ctx context.Context, nim string) (*entity.User, error) {
	var rows []entity.User
	if err := r.db.WithContext(ctx).Select(profileColumns).Where("nim = ?", nim).Limit(2).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, nil
	}
	return &rows[0], nil
}

// Meniru check_user_xp_level (migration-xp-fix-20261026.sql:54): XP ditetapkan dari sumber, nilai klien diabaikan.
// mult adalah ekspresi SQL prestige_mult yang sudah di-clamp ke NUMERIC(4,2), sama dengan mult di trigger.
func xpFromSources(mult string) string {
	return `(FLOOR((SELECT COALESCE(SUM(l.xp_reward), 0) FROM student_progress sp JOIN student_lessons l ON sp.lesson_id = l.id WHERE sp.nim = @nim AND sp.completed)::integer * ` + mult + `)::integer
 + (SELECT COALESCE(SUM(gh.xp_earned), 0) FROM game_history gh WHERE gh.nim = @nim)::integer)`
}

func clampMult(expr string) string {
	return "LEAST(GREATEST(COALESCE(" + expr + ", 1.00), 1.00), 1.50)::numeric(4,2)"
}

// Nama kolom berasal dari daftar column grant di usecase, jadi aman disisipkan ke SQL.
func assignments(cols map[string]any, skip ...string) (names, params []string, args map[string]any) {
	args = map[string]any{}
	for _, k := range slices.Sorted(maps.Keys(cols)) {
		if slices.Contains(skip, k) {
			continue
		}
		p := fmt.Sprintf("c%d", len(names))
		names = append(names, k)
		params = append(params, "@"+p)
		args[p] = cols[k]
	}
	return names, params, args
}

func (r *userRepository) Insert(ctx context.Context, cols map[string]any, recomputeXP bool) error {
	var skip []string
	if recomputeXP {
		skip = []string{"xp", "level", "prestige_mult"}
	}
	names, placeholders, args := assignments(cols, skip...)
	if recomputeXP {
		args["nim"] = cols["nim"]
		args["mult"] = cols["prestige_mult"]
		mult := clampMult("CAST(@mult AS numeric)")
		names = append(names, "prestige_mult", "xp", "level")
		placeholders = append(placeholders, mult, xpFromSources(mult), "public.calculate_level("+xpFromSources(mult)+")")
	}
	sql := "INSERT INTO users (" + strings.Join(names, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	return r.db.WithContext(ctx).Exec(sql, args).Error
}

type updateResult struct {
	Found     int64
	Updated   int64
	StudyTime *int64
}

// Satu statement: baris target, UPDATE, dan hasilnya dibaca bersama supaya tidak ada baca-hitung-tulis di Go.
func (r *userRepository) update(ctx context.Context, nim string, sets []string, args map[string]any, recomputeXP bool) (*updateResult, error) {
	args["nim"] = nim
	where := "nim = @nim"
	if recomputeXP {
		mult := clampMult("prestige_mult")
		xp := xpFromSources(mult)
		sets = append(sets, "xp = "+xp, "level = public.calculate_level("+xp+")")
		where += " AND prestige_mult IS NOT DISTINCT FROM " + mult
	}
	sql := `WITH target AS (SELECT 1 FROM users WHERE nim = @nim),
upd AS (UPDATE users SET ` + strings.Join(sets, ", ") + ` WHERE ` + where + ` RETURNING study_time)
SELECT (SELECT count(*) FROM target) AS found, (SELECT count(*) FROM upd) AS updated, (SELECT study_time FROM upd LIMIT 1) AS study_time`
	var res updateResult
	if err := r.db.WithContext(ctx).Raw(sql, args).Scan(&res).Error; err != nil {
		return nil, err
	}
	if recomputeXP && res.Found > 0 && res.Updated == 0 {
		return nil, ErrPrestigeGuard
	}
	return &res, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, nim string, cols map[string]any, recomputeXP bool) (int64, error) {
	var skip []string
	if recomputeXP {
		skip = []string{"xp", "level"}
	}
	names, params, args := assignments(cols, skip...)
	if len(names) == 0 && !recomputeXP {
		return 0, nil
	}
	sets := make([]string, len(names))
	for i, n := range names {
		sets[i] = n + " = " + params[i]
	}
	res, err := r.update(ctx, nim, sets, args, recomputeXP)
	if err != nil {
		return 0, err
	}
	return res.Updated, nil
}

// Menggantikan baca lalu tulis study_time di Layout.tsx:143-153 yang bisa saling menimpa antar tab.
func (r *userRepository) IncrementStudyTime(ctx context.Context, nim string, seconds int64, recomputeXP bool) (*int64, error) {
	res, err := r.update(ctx, nim, []string{"study_time = COALESCE(study_time, 0) + @seconds"}, map[string]any{"seconds": seconds}, recomputeXP)
	if err != nil {
		return nil, err
	}
	return res.StudyTime, nil
}

func (r *userRepository) ListAdmin(ctx context.Context, page, limit int, q string) ([]AdminUser, int64, error) {
	filter := noFilter
	if q != "" {
		like := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(q) + "%"
		filter = func(db *gorm.DB) *gorm.DB { return db.Where("nama ILIKE ? OR email ILIKE ?", like, like) }
	}
	return findPage[AdminUser](ctx, filter, r.db, "xp DESC, nim", page, limit)
}

func (r *userRepository) Delete(ctx context.Context, nim string) (int64, error) {
	res := r.db.WithContext(ctx).Where("nim = ?", nim).Delete(&entity.User{})
	return res.RowsAffected, res.Error
}

func (r *userRepository) XPLevel(ctx context.Context, nim string) (*XPLevel, error) {
	var rows []XPLevel
	err := r.db.WithContext(ctx).Model(&entity.User{}).Select("COALESCE(xp, 0) AS xp, COALESCE(level, 0) AS level").Where("nim = ?", nim).Limit(2).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rows[0], nil
}

// role in ('praktikan', 'mahasiswa') seperti StudentMonitoring.tsx:83; nim menjadi urutan stabil untuk halaman.
func (r *userRepository) ListStudents(ctx context.Context, page, limit int) ([]MonitoringStudent, int64, error) {
	filter := func(db *gorm.DB) *gorm.DB { return db.Where("role IN ?", []string{"praktikan", "mahasiswa"}) }
	return findPage[MonitoringStudent](ctx, filter, r.db, "nim", page, limit)
}

// Hasil nil berarti tidak tepat satu baris; checkPrestigeEligibility membaca .maybeSingle() yang gagal sebagai 0/null.
func (r *userRepository) PrestigeState(ctx context.Context, nim string) (*PrestigeState, error) {
	var rows []PrestigeState
	err := r.db.WithContext(ctx).Model(&entity.User{}).Select("COALESCE(prestige_count, 0) AS prestige_count, last_prestige_at").Where("nim = ?", nim).Limit(2).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, nil
	}
	return &rows[0], nil
}
