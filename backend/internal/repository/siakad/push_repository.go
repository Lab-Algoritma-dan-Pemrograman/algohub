package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

const pushSelectSQL = "SELECT user_id, endpoint, p256dh, auth FROM public.push_subscriptions"

type PushRepository interface {
	List(ctx context.Context, userFilter any) ([]entity.PushSubscription, error)
	DeleteByEndpoints(ctx context.Context, endpoints []string) error
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}

type pushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) PushRepository {
	return &pushRepository{db: db}
}

// userFilter nil berarti semua subscription, sama dengan query lama tanpa .eq().
func (r *pushRepository) List(ctx context.Context, userFilter any) ([]entity.PushSubscription, error) {
	q := r.db.WithContext(ctx)
	if userFilter == nil {
		q = q.Raw(pushSelectSQL)
	} else {
		q = q.Raw(pushSelectSQL+" WHERE user_id = ?", userFilter)
	}
	rows, err := q.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.PushSubscription
	for rows.Next() {
		var (
			userID              sql.NullInt64
			endpoint, key, auth sql.NullString
		)
		if err := rows.Scan(&userID, &endpoint, &key, &auth); err != nil {
			return nil, err
		}
		sub := entity.PushSubscription{Endpoint: endpoint.String, P256dh: key.String, Auth: auth.String}
		if userID.Valid {
			sub.UserID = &userID.Int64
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (r *pushRepository) DeleteByEndpoints(ctx context.Context, endpoints []string) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM public.push_subscriptions WHERE endpoint IN ?", endpoints).Error
}

func (r *pushRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM public.push_subscriptions WHERE endpoint = ?", endpoint).Error
}
