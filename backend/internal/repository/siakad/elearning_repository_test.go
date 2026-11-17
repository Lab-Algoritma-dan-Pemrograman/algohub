package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestElearningMaybeSingle(t *testing.T) {
	db, mock := newMockDB(t)
	const session = "SELECT to_json(s) FROM public.active_sessions AS s WHERE s.nim = $1 LIMIT 2"
	const overrides = "SELECT to_json(level_access_overrides) FROM public.users WHERE nim = $1 LIMIT 2"
	mock.ExpectQuery(session).WithArgs("1").WillReturnRows(
		sqlmock.NewRows([]string{"to_json"}).AddRow(`{"nim":"1","last_heartbeat":"2026-01-01T00:00:00+00:00"}`))
	mock.ExpectQuery(session).WithArgs("2").WillReturnRows(
		sqlmock.NewRows([]string{"to_json"}).AddRow(`{"last_heartbeat":null}`).AddRow(`{"last_heartbeat":null}`))
	mock.ExpectQuery(overrides).WithArgs("1").WillReturnRows(
		sqlmock.NewRows([]string{"to_json"}).AddRow(`{"level-2":"unlocked"}`))
	mock.ExpectQuery(overrides).WithArgs("2").WillReturnRows(
		sqlmock.NewRows([]string{"to_json"}).AddRow(`{}`).AddRow(`{}`))

	repo := NewElearningRepository(db)
	ctx := context.Background()
	if s, err := repo.ActiveSession(ctx, "1"); err != nil || s == nil || string(s.LastHeartbeat) != `"2026-01-01T00:00:00+00:00"` {
		t.Fatalf("session = %+v %v", s, err)
	}
	if s, err := repo.ActiveSession(ctx, "2"); err != nil || s != nil {
		t.Fatalf("dua baris harus nil: %+v %v", s, err)
	}
	if raw, err := repo.LevelOverrides(ctx, "1"); err != nil || string(raw) != `{"level-2":"unlocked"}` {
		t.Fatalf("overrides = %s %v", raw, err)
	}
	if raw, err := repo.LevelOverrides(ctx, "2"); err != nil || raw != nil {
		t.Fatalf("dua baris harus nil: %s %v", raw, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProgressExistsByNIM(t *testing.T) {
	db, mock := newMockDB(t)
	const q = "SELECT id FROM public.elearning_progress WHERE nim = $1 LIMIT 2"
	mock.ExpectQuery(q).WithArgs("1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(q).WithArgs("2").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	repo := NewProgressRepository(db)
	if ok, err := repo.ExistsByNIM(context.Background(), "1"); err != nil || !ok {
		t.Fatalf("satu baris: %v %v", ok, err)
	}
	if ok, err := repo.ExistsByNIM(context.Background(), "2"); err != nil || ok {
		t.Fatalf("dua baris: %v %v", ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
