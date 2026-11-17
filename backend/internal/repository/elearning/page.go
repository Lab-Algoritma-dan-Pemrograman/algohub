package repository

import (
	"context"

	"gorm.io/gorm"
)

// Satu pola untuk semua daftar berhalaman: total dihitung dari query yang sama sebelum offset/limit.
func findPage[T any](ctx context.Context, q func(*gorm.DB) *gorm.DB, db *gorm.DB, order string, page, limit int) (rows []T, total int64, err error) {
	if err := q(db.WithContext(ctx).Model(new(T))).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = q(db.WithContext(ctx).Model(new(T))).Order(order).Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func noFilter(db *gorm.DB) *gorm.DB { return db }
