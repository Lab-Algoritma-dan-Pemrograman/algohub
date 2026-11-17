package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "87654321"

var testHash = func() string {
	b, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}()

func loginRow(overrides map[string]any) json.RawMessage {
	row := map[string]any{
		"id": 7, "username": "202711001", "full_name": "Budi", "role": "mahasiswa", "password": testHash,
		"is_active": true, "shift": nil, "division": nil, "nim": "202711001", "assistant_code": nil,
		"class_code": "TI-1", "phone_number": "08", "major": "TI",
	}
	for k, v := range overrides {
		if v == "<hapus>" {
			delete(row, k)
			continue
		}
		row[k] = v
	}
	b, _ := json.Marshal(row)
	return b
}

func newAuth(row json.RawMessage) (*AuthUsecase, *fakeAuth) {
	auth := &fakeAuth{loginRaw: row, revoked: map[string]bool{}}
	return NewAuthUsecase(auth, NewTokens(testSecret)), auth
}

func TestLoginSuccess(t *testing.T) {
	uc, auth := newAuth(loginRow(map[string]any{"role": "asisten", "division": "<hapus>"}))
	body, token, f := uc.Login(context.Background(), "202711001", testPassword)
	if f != nil {
		t.Fatalf("Login: %+v", f)
	}
	if fmt.Sprint(auth.loginArgs) != "[202711001 87654321]" {
		t.Errorf("login_user args = %v", auth.loginArgs)
	}
	if body["token"] != token {
		t.Error("token di body harus sama dengan token cookie")
	}
	user := body["user"].(map[string]any)
	if _, ok := user["password"]; ok {
		t.Fatal("hash password bocor ke respons")
	}
	if user["class_code"] != "TI-1" || user["id"] != json.Number("7") {
		t.Errorf("user = %v", user)
	}

	claims, err := NewTokens(testSecret).Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	for _, k := range []string{"id", "username", "role", "nim", "assistant_code", "full_name", "jti", "iat", "exp"} {
		if _, ok := claims[k]; !ok {
			t.Errorf("klaim %s hilang", k)
		}
	}
	for _, k := range []string{"division", "class_code", "password", "is_active", "shift"} {
		if _, ok := claims[k]; ok {
			t.Errorf("klaim %s tidak boleh ada", k)
		}
	}
	iat, _ := claimFloat(claims["iat"])
	exp, _ := claimFloat(claims["exp"])
	if exp-iat != 3600 {
		t.Errorf("exp-iat = %v, ingin 3600", exp-iat)
	}
	if jti, _ := claims["jti"].(string); len(jti) != 36 {
		t.Errorf("jti = %q, ingin UUID", jti)
	}
}

func TestLoginFailures(t *testing.T) {
	cases := []struct {
		name               string
		username, password string
		row                json.RawMessage
		err                error
		status             int
		msg                string
	}{
		{"username kosong", "", "x", nil, nil, 400, "Username dan password wajib diisi."},
		{"password kosong", "u", "", nil, nil, 400, "Username dan password wajib diisi."},
		{"tidak ada baris", "u", testPassword, nil, nil, 401, "Username tidak ditemukan atau password salah!"},
		{"password salah", "u", "salah", loginRow(nil), nil, 401, "Username tidak ditemukan atau password salah!"},
		{"hash kosong", "u", testPassword, loginRow(map[string]any{"password": nil}), nil, 401, "Username tidak ditemukan atau password salah!"},
		{"nonaktif", "u", testPassword, loginRow(map[string]any{"is_active": false}), nil, 403, "Akun Anda sedang dinonaktifkan."},
		{"is_active null", "u", testPassword, loginRow(map[string]any{"is_active": nil}), nil, 403, "Akun Anda sedang dinonaktifkan."},
		{"pg sistem", "u", testPassword, nil, &pgconn.PgError{Message: `relation "users" does not exist`}, 400, "Terjadi kesalahan sistem. Silakan coba lagi."},
		{"pg bisnis", "u", testPassword, nil, &pgconn.PgError{Message: "Akun terkunci"}, 400, "Akun terkunci"},
		{"koneksi", "u", testPassword, nil, errors.New("dial tcp"), 500, "Internal server error"},
	}
	for _, tc := range cases {
		uc, auth := newAuth(tc.row)
		auth.loginErr = tc.err
		_, _, f := uc.Login(context.Background(), tc.username, tc.password)
		if f == nil || f.Status != tc.status || f.Body["error"] != tc.msg {
			t.Errorf("%s: ingin %d %q, dapat %+v", tc.name, tc.status, tc.msg, f)
		}
	}
}

func TestLoginShift(t *testing.T) {
	cases := []struct {
		name      string
		role      any
		shift     any
		active    string
		activeOK  bool
		wantErr   string
		wantDet   string
		wantQuery bool
	}{
		{"shift null tanpa cek", "mahasiswa", nil, "B", true, "", "", false},
		{"bukan mahasiswa", "asisten", "A", "B", true, "", "", false},
		{"alias praktikan tidak dicek", "praktikan", "A", "B", true, "", "", false},
		{"all", "mahasiswa", "A", "all", true, "", "", true},
		{"settings tidak tunggal = all", "mahasiswa", "A", "", false, "", "", true},
		{"active kosong = all", "mahasiswa", "A", "", true, "", "", true},
		{"none", "mahasiswa", "A", "none", true, "Akses Ditutup", "Maaf, akses login saat ini sedang ditutup untuk semua praktikan.", true},
		{"shift kosong", "mahasiswa", "", "B", true, "Akun Belum Aktif", "Mohon tunggu plotting shift dari Koordinator sebelum Anda bisa login.", true},
		{"shift ?", "mahasiswa", "?", "B", true, "Akun Belum Aktif", "Mohon tunggu plotting shift dari Koordinator sebelum Anda bisa login.", true},
		{"shift beda", "mahasiswa", "A", "B", true, "Akses Ditolak", "Maaf, akun Anda terdaftar di Shift A. Saat ini hanya Shift B yang diizinkan masuk.", true},
		{"shift sama", "mahasiswa", "B", "B", true, "", "", true},
	}
	for _, tc := range cases {
		uc, auth := newAuth(loginRow(map[string]any{"role": tc.role, "shift": tc.shift}))
		auth.shift, auth.shiftOK = tc.active, tc.activeOK
		_, _, f := uc.Login(context.Background(), "u", testPassword)
		switch {
		case tc.wantErr == "" && f != nil:
			t.Errorf("%s: harus lolos, dapat %+v", tc.name, f)
		case tc.wantErr != "" && (f == nil || f.Status != http.StatusForbidden || f.Body["error"] != tc.wantErr || f.Body["details"] != tc.wantDet):
			t.Errorf("%s: ingin 403 %q/%q, dapat %+v", tc.name, tc.wantErr, tc.wantDet, f)
		}
		if (auth.shiftCalls > 0) != tc.wantQuery {
			t.Errorf("%s: query system_settings = %d, ingin %v", tc.name, auth.shiftCalls, tc.wantQuery)
		}
	}
}

func TestLogout(t *testing.T) {
	uc, auth := newAuth(nil)
	exp := time.Now().Add(30 * time.Minute).Unix()
	token, _ := NewTokens(testSecret).Sign(map[string]any{"id": 42, "jti": "jti-1"}, 30*time.Minute)
	uc.Logout(context.Background(), token)
	if len(auth.revokes) != 1 {
		t.Fatalf("revoke_token dipanggil %d kali, ingin 1", len(auth.revokes))
	}
	got := auth.revokes[0]
	if got.jti != "jti-1" || got.userID != "42" || got.expiresAt.Location() != time.UTC {
		t.Errorf("revoke = %+v", got)
	}
	if d := got.expiresAt.Unix() - exp; d < -1 || d > 1 {
		t.Errorf("expiresAt = %v, ingin dekat %v", got.expiresAt, time.Unix(exp, 0))
	}

	skip := []string{
		"",
		"bukan-token",
		sign(t, map[string]any{"jti": "x"}),
		sign(t, map[string]any{"id": 0, "jti": "x"}),
		sign(t, map[string]any{"id": 1}),
	}
	for _, tok := range skip {
		uc, auth := newAuth(nil)
		uc.Logout(context.Background(), tok)
		if len(auth.revokes) != 0 {
			t.Errorf("token %q tidak boleh di-revoke", tok)
		}
	}
}

func TestVerify(t *testing.T) {
	uc, auth := newAuth(nil)
	auth.revoked["dicabut"] = true

	_, cleared, f := uc.Verify(context.Background(), "")
	if f == nil || f.Status != 401 || f.Body["error"] != "Unauthorized: Missing token" || cleared {
		t.Errorf("kosong: %+v clear=%v", f, cleared)
	}
	_, cleared, f = uc.Verify(context.Background(), "bukan-token")
	if f == nil || f.Body["error"] != "Unauthorized: Invalid or expired token" || !cleared {
		t.Errorf("invalid: %+v clear=%v", f, cleared)
	}
	_, cleared, f = uc.Verify(context.Background(), sign(t, map[string]any{"id": 1, "jti": "dicabut"}))
	if f == nil || f.Body["error"] != "Unauthorized: Session telah berakhir. Silakan login kembali." || !cleared {
		t.Errorf("dicabut: %+v clear=%v", f, cleared)
	}

	body, cleared, f := uc.Verify(context.Background(), sign(t, map[string]any{"id": 1, "role": "asisten", "jti": "aktif"}))
	if f != nil || cleared {
		t.Fatalf("valid: %+v clear=%v", f, cleared)
	}
	b, _ := json.Marshal(body)
	if string(b) != `{"user":{"id":1,"role":"asisten"}}` {
		t.Errorf("body = %s", b)
	}
}
