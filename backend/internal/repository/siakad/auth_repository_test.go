package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const loginSQL = `SELECT to_json(_r) FROM public.login_user(p_username => $1::text, p_password => $2::text) AS _r`

func TestLoginUser(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAuthRepository(db)

	mock.ExpectQuery(loginSQL).WithArgs("202711001", "pw").
		WillReturnRows(sqlmock.NewRows([]string{"r"}).AddRow([]byte(`{"id":1}`)))
	mock.ExpectQuery(loginSQL).WithArgs("tidak-ada", "pw").WillReturnRows(sqlmock.NewRows([]string{"r"}))

	got, err := repo.LoginUser(context.Background(), "202711001", "pw")
	if err != nil || string(got) != `{"id":1}` {
		t.Fatalf("LoginUser = %s, %v", got, err)
	}
	got, err = repo.LoginUser(context.Background(), "tidak-ada", "pw")
	if err != nil || got != nil {
		t.Fatalf("LoginUser tanpa baris = %s, %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestActiveShift(t *testing.T) {
	const q = `SELECT active_shift FROM public.system_settings LIMIT 2`
	cases := []struct {
		name   string
		rows   *sqlmock.Rows
		err    error
		want   string
		wantOK bool
	}{
		{"satu baris", sqlmock.NewRows([]string{"active_shift"}).AddRow("A"), nil, "A", true},
		{"null", sqlmock.NewRows([]string{"active_shift"}).AddRow(nil), nil, "", true},
		{"kosong", sqlmock.NewRows([]string{"active_shift"}), nil, "", false},
		{"dua baris", sqlmock.NewRows([]string{"active_shift"}).AddRow("A").AddRow("B"), nil, "", false},
		{"error", nil, errors.New("putus"), "", false},
	}
	for _, tc := range cases {
		db, mock := newMockDB(t)
		exp := mock.ExpectQuery(q)
		if tc.err != nil {
			exp.WillReturnError(tc.err)
		} else {
			exp.WillReturnRows(tc.rows)
		}
		got, ok := NewAuthRepository(db).ActiveShift(context.Background())
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("%s: (%q, %v), ingin (%q, %v)", tc.name, got, ok, tc.want, tc.wantOK)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("%s: %v", tc.name, err)
		}
	}
}

func TestTokenRevocation(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAuthRepository(db)
	exp := time.Date(2026, 11, 17, 3, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT coalesce(public.is_token_revoked(p_jti => $1::text), false)`).WithArgs("jti-1").
		WillReturnRows(sqlmock.NewRows([]string{"r"}).AddRow(true))
	mock.ExpectExec(`SELECT public.revoke_token(p_jti => $1::text, p_user_id => $2::bigint, p_expires_at => $3::timestamptz)`).
		WithArgs("jti-1", "42", exp).WillReturnResult(sqlmock.NewResult(0, 1))

	revoked, err := repo.IsTokenRevoked(context.Background(), "jti-1")
	if err != nil || !revoked {
		t.Fatalf("IsTokenRevoked = %v, %v", revoked, err)
	}
	if err := repo.RevokeToken(context.Background(), "jti-1", "42", exp); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
