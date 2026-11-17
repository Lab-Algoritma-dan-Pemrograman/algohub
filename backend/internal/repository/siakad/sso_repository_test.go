package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

func TestSSOHandshakeUserID(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	const q = "SELECT user_id FROM public.sso_handshakes WHERE code = $1 AND expires_at > $2 LIMIT 2"
	cases := []struct {
		name  string
		rows  []any
		found bool
		id    int64
	}{
		{"satu baris", []any{int64(9)}, true, 9},
		{"user_id null", []any{nil}, true, 0},
		{"tidak ada", nil, false, 0},
		{"lebih dari satu", []any{int64(1), int64(2)}, false, 0},
	}
	for _, tc := range cases {
		db, mock := newMockDB(t)
		rows := sqlmock.NewRows([]string{"user_id"})
		for _, v := range tc.rows {
			rows.AddRow(v)
		}
		mock.ExpectQuery(q).WithArgs("kode", now).WillReturnRows(rows)
		id, found, err := NewSSORepository(db).HandshakeUserID(context.Background(), "kode", now)
		if err != nil || found != tc.found || (tc.id != 0) != (id != nil) || (id != nil && *id != tc.id) {
			t.Errorf("%s: id %v found %v err %v", tc.name, id, found, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("%s: %v", tc.name, err)
		}
	}
}

func TestSSOUser(t *testing.T) {
	db, mock := newMockDB(t)
	const q = "SELECT id, nim, username, full_name, nama, role, class_code, major FROM public.users WHERE id = $1"
	mock.ExpectQuery(q).WithArgs(int64(9)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "nim", "username", "full_name", "nama", "role", "class_code", "major"}).
			AddRow(9, nil, "u", "Budi", nil, "asisten", "A", nil))
	mock.ExpectQuery(q).WithArgs(int64(10)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "nim", "username", "full_name", "nama", "role", "class_code", "major"}))
	repo := NewSSORepository(db)
	user, found, err := repo.User(context.Background(), 9)
	want := entity.SSOUser{Username: "u", FullName: "Budi", Role: "asisten", ClassCode: "A"}
	if err != nil || !found || user != want {
		t.Fatalf("User = %+v %v %v", user, found, err)
	}
	if _, found, err := repo.User(context.Background(), 10); err != nil || found {
		t.Fatalf("tidak ada: %v %v", found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSSODeleteAndSchedules(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec("DELETE FROM public.sso_handshakes WHERE code = $1").WithArgs("kode").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT s.class_code, s.major FROM public.group_members gm LEFT JOIN public.schedules s ON s.id = gm.schedule_id WHERE gm.student_id = $1 LIMIT 1").
		WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"class_code", "major"}).AddRow("K1", nil))
	mock.ExpectQuery("SELECT s.class_code, s.major FROM public.group_assistants ga LEFT JOIN public.schedules s ON s.id = ga.schedule_id WHERE ga.assistant_id = $1 LIMIT 1").
		WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"class_code", "major"}))
	mock.ExpectQuery("SELECT s.class_code, s.major FROM public.group_assistants ga LEFT JOIN public.schedules s ON s.id = ga.schedule_id WHERE ga.assistant_id = $1 LIMIT 1").
		WithArgs(int64(10)).WillReturnError(errors.New("relation does not exist"))

	repo := NewSSORepository(db)
	if err := repo.DeleteHandshake(context.Background(), "kode"); err != nil {
		t.Fatal(err)
	}
	if s, err := repo.StudentSchedule(context.Background(), 9); err != nil || s != (entity.ScheduleClass{ClassCode: "K1"}) {
		t.Fatalf("student = %+v %v", s, err)
	}
	if s, err := repo.AssistantSchedule(context.Background(), 9); err != nil || s != (entity.ScheduleClass{}) {
		t.Fatalf("assistant kosong = %+v %v", s, err)
	}
	if _, err := repo.AssistantSchedule(context.Background(), 10); err == nil {
		t.Fatal("error query harus diteruskan")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
