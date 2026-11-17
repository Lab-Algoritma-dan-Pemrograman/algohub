package repository

import (
	"context"
	"maps"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// Nama tabel dan kolom di helper ini selalu literal dari repository atau daftar kolom tetap di usecase, jadi aman disisipkan ke SQL.

func placeholders(n int) string {
	return "(" + strings.TrimSuffix(strings.Repeat("?, ", n), ", ") + ")"
}

// Satu INSERT multi-baris menggantikan .insert(array) PostgREST.
func insertRows(db *gorm.DB, table string, cols []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}
	values := make([]string, len(rows))
	args := make([]any, 0, len(rows)*len(cols))
	for i, row := range rows {
		values[i] = placeholders(len(cols))
		args = append(args, row...)
	}
	return db.Exec("INSERT INTO "+table+" ("+strings.Join(cols, ", ")+") VALUES "+strings.Join(values, ", "), args...).Error
}

// Setara .insert([...]).select().single() yang hanya memakai id hasil insert.
func insertReturningID(ctx context.Context, db *gorm.DB, table string, cols map[string]any) (string, error) {
	names := slices.Sorted(maps.Keys(cols))
	args := make([]any, len(names))
	for i, n := range names {
		args[i] = cols[n]
	}
	var row struct{ ID string }
	err := db.WithContext(ctx).Raw("INSERT INTO "+table+" ("+strings.Join(names, ", ")+") VALUES "+placeholders(len(names))+" RETURNING id::text AS id", args...).Scan(&row).Error
	return row.ID, err
}

func updateByID(ctx context.Context, db *gorm.DB, table, id string, cols map[string]any) (int64, error) {
	if len(cols) == 0 {
		return 0, nil
	}
	res := db.WithContext(ctx).Table(table).Where("id = ?", id).Updates(cols)
	return res.RowsAffected, res.Error
}

func deleteByID(ctx context.Context, db *gorm.DB, table, id string) (int64, error) {
	res := db.WithContext(ctx).Exec("DELETE FROM "+table+" WHERE id = ?", id)
	return res.RowsAffected, res.Error
}
