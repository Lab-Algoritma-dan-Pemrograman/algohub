package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

const (
	ssoHandshakeSQL      = "SELECT user_id FROM public.sso_handshakes WHERE code = ? AND expires_at > ? LIMIT 2"
	ssoUserSQL           = "SELECT id, nim, username, full_name, nama, role, class_code, major FROM public.users WHERE id = ?"
	ssoDeleteSQL         = "DELETE FROM public.sso_handshakes WHERE code = ?"
	studentScheduleSQL   = "SELECT s.class_code, s.major FROM public.group_members gm LEFT JOIN public.schedules s ON s.id = gm.schedule_id WHERE gm.student_id = ? LIMIT 1"
	assistantScheduleSQL = "SELECT s.class_code, s.major FROM public.group_assistants ga LEFT JOIN public.schedules s ON s.id = ga.schedule_id WHERE ga.assistant_id = ? LIMIT 1"
)

type SSORepository interface {
	HandshakeUserID(ctx context.Context, code string, now time.Time) (userID *int64, found bool, err error)
	User(ctx context.Context, id int64) (entity.SSOUser, bool, error)
	DeleteHandshake(ctx context.Context, code string) error
	StudentSchedule(ctx context.Context, userID int64) (entity.ScheduleClass, error)
	AssistantSchedule(ctx context.Context, userID int64) (entity.ScheduleClass, error)
}

type ssoRepository struct {
	db *gorm.DB
}

func NewSSORepository(db *gorm.DB) SSORepository {
	return &ssoRepository{db: db}
}

// Meniru .maybeSingle(): handshake hanya sah kalau tepat satu baris yang cocok.
func (r *ssoRepository) HandshakeUserID(ctx context.Context, code string, now time.Time) (userID *int64, found bool, err error) {
	rows, err := r.db.WithContext(ctx).Raw(ssoHandshakeSQL, code, now).Rows()
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var (
		raw   sql.NullInt64
		count int
	)
	for rows.Next() {
		if err := rows.Scan(&raw); err != nil {
			return nil, false, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	if count != 1 {
		return nil, false, nil
	}
	if !raw.Valid {
		return nil, true, nil
	}
	return &raw.Int64, true, nil
}

func (r *ssoRepository) User(ctx context.Context, id int64) (entity.SSOUser, bool, error) {
	var (
		rowID                                             sql.NullInt64
		nim, username, fullName, nama, role, class, major sql.NullString
	)
	err := r.db.WithContext(ctx).Raw(ssoUserSQL, id).Row().
		Scan(&rowID, &nim, &username, &fullName, &nama, &role, &class, &major)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.SSOUser{}, false, nil
	}
	if err != nil {
		return entity.SSOUser{}, false, err
	}
	return entity.SSOUser{
		NIM: nim.String, Username: username.String, FullName: fullName.String, Nama: nama.String,
		Role: role.String, ClassCode: class.String, Major: major.String,
	}, true, nil
}

func (r *ssoRepository) DeleteHandshake(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Exec(ssoDeleteSQL, code).Error
}

func (r *ssoRepository) StudentSchedule(ctx context.Context, userID int64) (entity.ScheduleClass, error) {
	return r.firstSchedule(ctx, studentScheduleSQL, userID)
}

func (r *ssoRepository) AssistantSchedule(ctx context.Context, userID int64) (entity.ScheduleClass, error) {
	return r.firstSchedule(ctx, assistantScheduleSQL, userID)
}

// Tidak ada baris jadwal bukan error: kelas dan jurusan dari tabel users tetap dipakai.
func (r *ssoRepository) firstSchedule(ctx context.Context, query string, userID int64) (entity.ScheduleClass, error) {
	var class, major sql.NullString
	err := r.db.WithContext(ctx).Raw(query, userID).Row().Scan(&class, &major)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ScheduleClass{}, nil
	}
	if err != nil {
		return entity.ScheduleClass{}, err
	}
	return entity.ScheduleClass{ClassCode: class.String, Major: major.String}, nil
}
