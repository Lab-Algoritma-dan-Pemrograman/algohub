package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPushList(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPushRepository(db)
	cols := []string{"user_id", "endpoint", "p256dh", "auth"}

	mock.ExpectQuery(pushSelectSQL).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(7), "https://e/1", "k1", "a1").AddRow(nil, "https://e/2", "k2", "a2"))
	mock.ExpectQuery(pushSelectSQL + " WHERE user_id = $1").WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(7), "https://e/1", "k1", "a1"))
	mock.ExpectQuery(pushSelectSQL + " WHERE user_id = $1").WithArgs("abc").WillReturnError(errors.New("invalid input syntax"))

	all, err := repo.List(context.Background(), nil)
	if err != nil || len(all) != 2 || *all[0].UserID != 7 || all[1].UserID != nil || all[1].P256dh != "k2" {
		t.Fatalf("List(nil) = %+v, %v", all, err)
	}
	one, err := repo.List(context.Background(), int64(7))
	if err != nil || len(one) != 1 || one[0].Auth != "a1" {
		t.Fatalf("List(7) = %+v, %v", one, err)
	}
	if _, err := repo.List(context.Background(), "abc"); err == nil {
		t.Fatal("error query harus diteruskan")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPushDelete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewPushRepository(db)

	mock.ExpectExec("DELETE FROM public.push_subscriptions WHERE endpoint IN ($1,$2)").
		WithArgs("https://e/1", "https://e/2").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM public.push_subscriptions WHERE endpoint = $1").
		WithArgs("https://e/3").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteByEndpoints(context.Background(), []string{"https://e/1", "https://e/2"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByEndpoint(context.Background(), "https://e/3"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWASettings(t *testing.T) {
	const q = "SELECT wa_gateway_provider, wa_gateway_token, wa_auto_notify_enabled FROM public.system_settings LIMIT 1"
	cols := []string{"wa_gateway_provider", "wa_gateway_token", "wa_auto_notify_enabled"}
	db, mock := newMockDB(t)
	repo := NewNotifyRepository(db)

	mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows(cols).AddRow("wablas", "tok", false))
	mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows(cols).AddRow(nil, nil, nil))
	mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows(cols))

	s, err := repo.WASettings(context.Background())
	if err != nil || s.Provider != "wablas" || s.Token != "tok" || s.AutoNotifyEnabled == nil || *s.AutoNotifyEnabled {
		t.Fatalf("WASettings = %+v, %v", s, err)
	}
	s, err = repo.WASettings(context.Background())
	if err != nil || s.Provider != "" || s.AutoNotifyEnabled != nil {
		t.Fatalf("WASettings NULL = %+v, %v", s, err)
	}
	s, err = repo.WASettings(context.Background())
	if err != nil || s.AutoNotifyEnabled != nil {
		t.Fatalf("WASettings kosong = %+v, %v", s, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserPhone(t *testing.T) {
	const q = "SELECT " + userPhoneColumn + " FROM public.users WHERE id = $1"
	db, mock := newMockDB(t)
	repo := NewNotifyRepository(db)

	mock.ExpectQuery(q).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"phone"}).AddRow("0812"))
	mock.ExpectQuery(q).WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"phone"}))

	if p, err := repo.UserPhone(context.Background(), 7); err != nil || p != "0812" {
		t.Fatalf("UserPhone(7) = %q, %v", p, err)
	}
	if p, err := repo.UserPhone(context.Background(), 8); err != nil || p != "" {
		t.Fatalf("UserPhone(8) = %q, %v", p, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserPhoneColumnFallsBack(t *testing.T) {
	if strings.Contains(userPhoneColumn, "public.") || strings.Contains(userPhoneColumn, "phone,") {
		t.Fatalf("kolom harus nomor_hp lalu phone_number: %s", userPhoneColumn)
	}
	for _, col := range []string{"nomor_hp", "phone_number"} {
		if !strings.Contains(userPhoneColumn, col) {
			t.Fatalf("kolom %s tidak dipakai: %s", col, userPhoneColumn)
		}
	}
	if strings.Contains(userPhoneColumn, "SELECT") || strings.Contains(userPhoneColumn, ";") {
		t.Fatalf("kolom tidak boleh memuat SQL: %s", userPhoneColumn)
	}
}

func TestTodayAssignments(t *testing.T) {
	db, mock := newMockDB(t)
	cols := []string{"user_id", "start_time", "role", "task_role", "activity_name"}
	mock.ExpectQuery(strings.Replace(todayAssignmentsSQL, "?", "$1", 1)).WithArgs("2026-11-17").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(7), "08:00:00", "asisten", "PJ", "Praktikum AP").
			AddRow(nil, nil, nil, nil, nil))

	got, err := NewShiftRepository(db).TodayAssignments(context.Background(), "2026-11-17")
	if err != nil || len(got) != 2 {
		t.Fatalf("TodayAssignments = %+v, %v", got, err)
	}
	if *got[0].UserID != 7 || got[0].StartTime != "08:00:00" || got[0].UserRole != "asisten" ||
		got[0].TaskRole != "PJ" || got[0].ActivityName != "Praktikum AP" {
		t.Errorf("baris 1 = %+v", got[0])
	}
	if got[1].UserID != nil || got[1].StartTime != "" {
		t.Errorf("baris NULL = %+v", got[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPingers(t *testing.T) {
	siakadDB, siakadMock := newMockDB(t)
	elearningDB, elearningMock := newMockDB(t)

	siakadMock.ExpectExec("SELECT public.get_public_settings()").WillReturnResult(sqlmock.NewResult(0, 1))
	elearningMock.ExpectExec("SELECT id FROM lessons LIMIT 1").WillReturnError(errors.New("paused"))

	if err := NewSiakadPinger(siakadDB).Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := NewElearningPinger(elearningDB).Ping(context.Background()); err == nil {
		t.Fatal("error ping elearning harus diteruskan")
	}
	if err := siakadMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := elearningMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
