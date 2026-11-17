package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AuthRepository interface {
	LoginUser(ctx context.Context, username, password string) (json.RawMessage, error)
	ActiveShift(ctx context.Context) (string, bool)
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
	RevokeToken(ctx context.Context, jti, userID string, expiresAt time.Time) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) LoginUser(ctx context.Context, username, password string) (json.RawMessage, error) {
	var out []byte
	err := r.db.WithContext(ctx).
		Raw("SELECT to_json(_r) FROM public.login_user(p_username => ?::text, p_password => ?::text) AS _r", username, password).
		Row().Scan(&out)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Meniru .maybeSingle(): nilai hanya dipakai kalau tabel berisi tepat satu baris.
func (r *authRepository) ActiveShift(ctx context.Context) (string, bool) {
	rows, err := r.db.WithContext(ctx).Raw("SELECT active_shift FROM public.system_settings LIMIT 2").Rows()
	if err != nil {
		return "", false
	}
	defer rows.Close()

	var (
		shift sql.NullString
		count int
	)
	for rows.Next() {
		if err := rows.Scan(&shift); err != nil {
			return "", false
		}
		count++
	}
	if rows.Err() != nil || count != 1 {
		return "", false
	}
	return shift.String, true
}

func (r *authRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var revoked bool
	err := r.db.WithContext(ctx).
		Raw("SELECT coalesce(public.is_token_revoked(p_jti => ?::text), false)", jti).
		Row().Scan(&revoked)
	return revoked, err
}

func (r *authRepository) RevokeToken(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).
		Exec("SELECT public.revoke_token(p_jti => ?::text, p_user_id => ?::bigint, p_expires_at => ?::timestamptz)", jti, userID, expiresAt).
		Error
}
