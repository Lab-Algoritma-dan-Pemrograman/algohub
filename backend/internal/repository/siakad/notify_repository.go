package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

// DB inti tidak punya kolom `phone`; nomor tersimpan di `nomor_hp` (130 baris) dan `phone_number` (13 baris).
const userPhoneColumn = "coalesce(nullif(trim(nomor_hp), ''), nullif(trim(phone_number), ''))"

type NotifyRepository interface {
	WASettings(ctx context.Context) (entity.WASettings, error)
	UserPhone(ctx context.Context, id int64) (string, error)
}

type notifyRepository struct {
	db *gorm.DB
}

func NewNotifyRepository(db *gorm.DB) NotifyRepository {
	return &notifyRepository{db: db}
}

// Tabel kosong bukan error: kode lama memakai {} dan jatuh ke default provider/token.
func (r *notifyRepository) WASettings(ctx context.Context) (entity.WASettings, error) {
	var (
		provider, token sql.NullString
		enabled         sql.NullBool
	)
	err := r.db.WithContext(ctx).
		Raw("SELECT wa_gateway_provider, wa_gateway_token, wa_auto_notify_enabled FROM public.system_settings LIMIT 1").
		Row().Scan(&provider, &token, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.WASettings{}, nil
	}
	if err != nil {
		return entity.WASettings{}, err
	}
	s := entity.WASettings{Provider: provider.String, Token: token.String}
	if enabled.Valid {
		s.AutoNotifyEnabled = &enabled.Bool
	}
	return s, nil
}

func (r *notifyRepository) UserPhone(ctx context.Context, id int64) (string, error) {
	var phone sql.NullString
	err := r.db.WithContext(ctx).Raw("SELECT "+userPhoneColumn+" FROM public.users WHERE id = ?", id).Row().Scan(&phone)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return phone.String, err
}
