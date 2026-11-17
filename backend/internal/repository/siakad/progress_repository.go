package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

var progressColumns = []string{
	"nim", "student_name", "completed_lessons", "lessons_completed", "total_lessons", "completion_percentage",
	"is_completed", "completed_levels", "current_level", "last_accessed_at", "updated_at",
}

// Baris dikirim sebagai satu parameter JSON dan dipetakan ke tipe kolom oleh json_populate_record,
// sama seperti PostgREST mengonversi body upsert.
const progressRecord = "json_populate_record(NULL::public.elearning_progress, ?::json)"

var (
	progressInsertSQL = "INSERT INTO public.elearning_progress (" + strings.Join(progressColumns, ", ") + ") SELECT " +
		strings.Join(progressColumns, ", ") + " FROM " + progressRecord
	progressUpsertSQL = progressInsertSQL + " ON CONFLICT (nim) DO UPDATE SET " + progressAssignments("EXCLUDED")
	progressUpdateSQL = "UPDATE public.elearning_progress AS t SET " + progressAssignments("r") +
		" FROM " + progressRecord + " AS r WHERE t.nim = r.nim"
	progressExistsSQL = "SELECT id FROM public.elearning_progress WHERE nim = ? LIMIT 2"
)

func progressAssignments(source string) string {
	sets := make([]string, 0, len(progressColumns)-1)
	for _, c := range progressColumns[1:] {
		sets = append(sets, c+" = "+source+"."+c)
	}
	return strings.Join(sets, ", ")
}

// Menulis ke DB siakad: tabel elearning_progress ada di project siakad, bukan elearning (K-6).
type ProgressRepository interface {
	Upsert(ctx context.Context, row *entity.ElearningProgressRow) error
	ExistsByNIM(ctx context.Context, nim string) (bool, error)
	Update(ctx context.Context, row *entity.ElearningProgressRow) error
	Insert(ctx context.Context, row *entity.ElearningProgressRow) error
}

type progressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) ProgressRepository {
	return &progressRepository{db: db}
}

func (r *progressRepository) Upsert(ctx context.Context, row *entity.ElearningProgressRow) error {
	return r.exec(ctx, progressUpsertSQL, row)
}

func (r *progressRepository) Update(ctx context.Context, row *entity.ElearningProgressRow) error {
	return r.exec(ctx, progressUpdateSQL, row)
}

func (r *progressRepository) Insert(ctx context.Context, row *entity.ElearningProgressRow) error {
	return r.exec(ctx, progressInsertSQL, row)
}

// Meniru .maybeSingle(): baris dianggap ada hanya kalau tepat satu.
func (r *progressRepository) ExistsByNIM(ctx context.Context, nim string) (bool, error) {
	rows, err := r.db.WithContext(ctx).Raw(progressExistsSQL, nim).Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id sql.RawBytes
		if err := rows.Scan(&id); err != nil {
			return false, err
		}
		count++
	}
	return count == 1, rows.Err()
}

func (r *progressRepository) exec(ctx context.Context, query string, row *entity.ElearningProgressRow) error {
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec(query, string(raw)).Error
}
