package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

const testSecret = "rahasia-uji"

type rpcCall struct {
	fn     string
	params map[string]json.RawMessage
}

type fakeRPC struct {
	calls  []rpcCall
	result json.RawMessage
	err    error
}

func (f *fakeRPC) Call(_ context.Context, fn string, params map[string]json.RawMessage) (json.RawMessage, error) {
	f.calls = append(f.calls, rpcCall{fn: fn, params: params})
	if f.err != nil {
		return nil, f.err
	}
	if f.result == nil {
		return json.RawMessage(`{"ok":true}`), nil
	}
	return f.result, nil
}

type revokeCall struct {
	jti, userID string
	expiresAt   time.Time
}

type fakeAuth struct {
	loginRaw   json.RawMessage
	loginErr   error
	loginArgs  []string
	shift      string
	shiftOK    bool
	shiftCalls int
	revoked    map[string]bool
	revokedErr error
	revokes    []revokeCall
}

func (f *fakeAuth) LoginUser(_ context.Context, username, password string) (json.RawMessage, error) {
	f.loginArgs = []string{username, password}
	return f.loginRaw, f.loginErr
}

func (f *fakeAuth) ActiveShift(context.Context) (string, bool) {
	f.shiftCalls++
	return f.shift, f.shiftOK
}

func (f *fakeAuth) IsTokenRevoked(_ context.Context, jti string) (bool, error) {
	return f.revoked[jti], f.revokedErr
}

func (f *fakeAuth) RevokeToken(_ context.Context, jti, userID string, expiresAt time.Time) error {
	f.revokes = append(f.revokes, revokeCall{jti, userID, expiresAt})
	return nil
}

func sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	token, err := NewTokens(testSecret).Sign(claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return token
}

func newGateway() (*RPCUsecase, *fakeRPC, *fakeAuth) {
	rpc := &fakeRPC{}
	auth := &fakeAuth{revoked: map[string]bool{}}
	return NewRPCUsecase(rpc, auth, NewTokens(testSecret)), rpc, auth
}

var roleAliasCases = map[string]string{
	"praktikan":   "praktikan",
	"mahasiswa":   "praktikan",
	"mhs":         "praktikan",
	"penyewa":     "penyewa",
	"peminjam":    "penyewa",
	"asisten":     "asisten",
	"admin":       "asisten",
	"koordinator": "koordinator",
	"kordas":      "koordinator",
	"korda":       "koordinator",
	"coordinator": "koordinator",
}

func allRPCNames() []string {
	var names []string
	for _, list := range rolePermissions {
		names = append(names, list...)
	}
	names = append(names, publicRPCs...)
	names = append(names, "login_user", "drop_everything", "get_users")
	slices.Sort(names)
	return slices.Compact(names)
}

func TestWhitelistShape(t *testing.T) {
	want := map[string]int{"praktikan": 22, "penyewa": 5, "asisten": 67, "koordinator": 76}
	for role, n := range want {
		list := rolePermissions[role]
		if len(list) != n {
			t.Errorf("%s: %d RPC, ingin %d", role, len(list), n)
		}
		sorted := slices.Sorted(slices.Values(list))
		if len(slices.Compact(sorted)) != len(list) {
			t.Errorf("%s: whitelist berisi duplikat", role)
		}
		if slices.Contains(list, "login_user") {
			t.Errorf("%s: login_user tidak boleh di whitelist", role)
		}
	}
	if len(rolePermissions) != len(want) {
		t.Errorf("role di whitelist = %d, ingin %d", len(rolePermissions), len(want))
	}
	if len(publicRPCs) != 5 || slices.Contains(publicRPCs, "login_user") {
		t.Errorf("publicRPCs = %v", publicRPCs)
	}
	for _, fn := range praktikanRPCs {
		if !slices.Contains(asistenRPCs, fn) {
			t.Errorf("%s ada di praktikan tetapi tidak di asisten", fn)
		}
	}
	if !slices.Equal(rolePermissions["koordinator"][:67], asistenRPCs) {
		t.Error("koordinator harus diawali seluruh whitelist asisten")
	}
	pinned := []struct {
		role, fn string
		want     bool
	}{
		{"praktikan", "get_users_secure", false},
		{"praktikan", "save_push_subscription_secure", true},
		{"penyewa", "get_schedules_secure", false},
		{"penyewa", "submit_rental_request_secure", true},
		{"asisten", "admin_create_user", false},
		{"asisten", "get_audit_logs_secure", true},
		{"koordinator", "admin_create_user", true},
		{"koordinator", "sync_class_rosters_secure", true},
	}
	for _, p := range pinned {
		if slices.Contains(rolePermissions[p.role], p.fn) != p.want {
			t.Errorf("%s × %s: ingin %v", p.role, p.fn, p.want)
		}
	}
}

func TestGatewayRoleMatrix(t *testing.T) {
	names := allRPCNames()
	for raw, canon := range roleAliasCases {
		token := sign(t, map[string]any{"id": 7, "role": raw, "jti": "j-" + raw})
		for _, fn := range names {
			uc, rpc, _ := newGateway()
			_, f := uc.Execute(context.Background(), fn, token, nil)
			allowed := slices.Contains(publicRPCs, fn) || slices.Contains(rolePermissions[canon], fn)
			if allowed {
				if f != nil || len(rpc.calls) != 1 || rpc.calls[0].fn != fn {
					t.Errorf("%s × %s: harus diizinkan, failure %+v, calls %v", raw, fn, f, rpc.calls)
				}
				continue
			}
			want := fmt.Sprintf("Forbidden: Peran Anda (%s) tidak diizinkan memanggil fungsi ini", canon)
			if f == nil || f.Status != http.StatusForbidden || f.Body["error"] != want || len(rpc.calls) != 0 {
				t.Errorf("%s × %s: ingin 403 %q, dapat %+v", raw, fn, want, f)
			}
		}
	}
}

func TestGatewayNonStringRoleIsUnauthorized(t *testing.T) {
	for _, role := range []any{5, true, []any{"asisten"}} {
		uc, rpc, _ := newGateway()
		_, f := uc.Execute(context.Background(), "get_user_profile", sign(t, map[string]any{"id": 1, "role": role}), nil)
		if f == nil || f.Status != http.StatusUnauthorized || f.Body["error"] != "Unauthorized: Invalid or expired token" || len(rpc.calls) != 0 {
			t.Errorf("role %v: %+v", role, f)
		}
	}
}

func TestErrTextHidesConnectionDetails(t *testing.T) {
	err := errors.New("failed to connect to `user=postgres.abc database=postgres`: db.example.com:6543")
	if got := errText(err); got != "Internal server error" {
		t.Fatalf("errText = %q", got)
	}
}
func TestGatewayInvalidRole(t *testing.T) {
	for _, role := range []string{"superadmin", "", "dosen"} {
		uc, rpc, _ := newGateway()
		_, f := uc.Execute(context.Background(), "get_user_profile", sign(t, map[string]any{"id": 1, "role": role}), nil)
		if f == nil || f.Status != http.StatusForbidden || f.Body["error"] != "Forbidden: Peran pengguna tidak valid" || len(rpc.calls) != 0 {
			t.Errorf("role %q: %+v", role, f)
		}
	}
}

func TestGatewayPublicWithoutToken(t *testing.T) {
	for _, fn := range publicRPCs {
		for _, token := range []string{"", sign(t, map[string]any{"id": 7, "role": "mahasiswa"})} {
			uc, rpc, _ := newGateway()
			params := map[string]json.RawMessage{"p_user_id": json.RawMessage(`99`), "p_caller_id": json.RawMessage(`98`)}
			if _, f := uc.Execute(context.Background(), fn, token, params); f != nil {
				t.Fatalf("%s: %+v", fn, f)
			}
			got := rpc.calls[0].params
			if string(got["p_user_id"]) != "99" || string(got["p_caller_id"]) != "98" {
				t.Errorf("%s: RPC publik tidak boleh diinjeksi, dapat %v", fn, got)
			}
		}
	}
}

func TestGatewayLoginUserBlocked(t *testing.T) {
	uc, rpc, _ := newGateway()
	_, f := uc.Execute(context.Background(), "login_user", "", nil)
	if f == nil || f.Status != http.StatusUnauthorized || f.Body["error"] != "Unauthorized: Missing session token" {
		t.Fatalf("tanpa token: %+v", f)
	}
	for raw := range roleAliasCases {
		_, f := uc.Execute(context.Background(), "login_user", sign(t, map[string]any{"id": 1, "role": raw}), nil)
		if f == nil || f.Status != http.StatusForbidden {
			t.Errorf("%s: login_user harus 403, dapat %+v", raw, f)
		}
	}
	if len(rpc.calls) != 0 {
		t.Fatalf("login_user tidak boleh sampai DB: %v", rpc.calls)
	}
}

func TestGatewayTokenChecks(t *testing.T) {
	expired, _ := NewTokens(testSecret).Sign(map[string]any{"id": 1, "role": "asisten"}, -time.Minute)
	wrong, _ := NewTokens("lain").Sign(map[string]any{"id": 1, "role": "asisten"}, time.Hour)
	cases := []struct {
		name       string
		token      string
		revokedErr error
		status     int
		msg        string
	}{
		{"kosong", "", nil, 401, "Unauthorized: Missing session token"},
		{"kedaluwarsa", expired, nil, 401, "Unauthorized: Invalid or expired token"},
		{"secret salah", wrong, nil, 401, "Unauthorized: Invalid or expired token"},
		{"dicabut", sign(t, map[string]any{"id": 1, "role": "asisten", "jti": "dicabut"}), nil, 401, "Unauthorized: Session telah berakhir. Silakan login kembali."},
		{"lookup error", sign(t, map[string]any{"id": 1, "role": "asisten", "jti": "dicabut"}), errors.New("putus"), 0, ""},
		{"tanpa jti", sign(t, map[string]any{"id": 1, "role": "asisten"}), nil, 0, ""},
	}
	for _, tc := range cases {
		uc, rpc, auth := newGateway()
		auth.revoked["dicabut"] = true
		auth.revokedErr = tc.revokedErr
		_, f := uc.Execute(context.Background(), "get_users_secure", tc.token, nil)
		if tc.status == 0 {
			if f != nil || len(rpc.calls) != 1 {
				t.Errorf("%s: harus lolos, dapat %+v", tc.name, f)
			}
			continue
		}
		if f == nil || f.Status != tc.status || f.Body["error"] != tc.msg || len(rpc.calls) != 0 {
			t.Errorf("%s: ingin %d %q, dapat %+v", tc.name, tc.status, tc.msg, f)
		}
	}
}

func TestGatewayIdentityInjection(t *testing.T) {
	raw := func(s string) json.RawMessage { return json.RawMessage(s) }
	cases := []struct {
		name   string
		claims map[string]any
		fn     string
		params map[string]json.RawMessage
		want   map[string]string
	}{
		{"praktikan get_user_profile", map[string]any{"id": 7, "role": "mahasiswa"}, "get_user_profile",
			map[string]json.RawMessage{"p_caller_id": raw(`1`), "p_target_id": raw(`2`)},
			map[string]string{"p_caller_id": "7", "p_target_id": "7"}},
		{"asisten get_user_profile orang lain", map[string]any{"id": 7, "role": "asisten"}, "get_user_profile",
			map[string]json.RawMessage{"p_target_id": raw(`2`)},
			map[string]string{"p_caller_id": "7", "p_target_id": "2"}},
		{"asisten get_user_profile diri", map[string]any{"id": 7, "role": "admin"}, "get_user_profile",
			nil, map[string]string{"p_caller_id": "7", "p_target_id": "7"}},
		{"praktikan semua id", map[string]any{"id": 7, "role": "praktikan"}, "upsert_attendance_log_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`), "p_target_user_id": raw(`1`), "p_target_id": raw(`1`), "p_viewer_id": raw(`1`), "p_student_id": raw(`1`), "p_note": raw(`"x"`)},
			map[string]string{"p_user_id": "7", "p_target_user_id": "7", "p_target_id": "7", "p_viewer_id": "7", "p_student_id": "7", "p_note": `"x"`}},
		{"penyewa p_user_id", map[string]any{"id": 7, "role": "peminjam"}, "get_my_rentals_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": "7"}},
		{"asisten tetap pakai id klien", map[string]any{"id": 7, "role": "asisten"}, "upsert_attendance_log_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`), "p_target_id": raw(`2`), "p_viewer_id": raw(`3`)},
			map[string]string{"p_user_id": "1", "p_target_id": "2", "p_viewer_id": "7"}},
		{"koordinator is_pj_absen_today", map[string]any{"id": 7, "role": "kordas"}, "is_pj_absen_today",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": "7"}},
		{"asisten check_pj_absen_today", map[string]any{"id": 7, "role": "asisten"}, "check_pj_absen_today",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": "7"}},
		{"asisten submit_rental", map[string]any{"id": 7, "role": "asisten"}, "submit_rental_request_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": "7"}},
		{"request_reschedule selalu caller", map[string]any{"id": 7, "role": "asisten"}, "request_reschedule_secure",
			nil, map[string]string{"p_caller_id": "7"}},
		{"asisten update_password", map[string]any{"id": 7, "role": "asisten"}, "update_password",
			map[string]json.RawMessage{"p_target_id": raw(`1`)}, map[string]string{"p_target_id": "7"}},
		{"koordinator update_password orang lain", map[string]any{"id": 7, "role": "koordinator"}, "update_password",
			map[string]json.RawMessage{"p_target_id": raw(`1`)}, map[string]string{"p_target_id": "1"}},
		{"id string", map[string]any{"id": "7", "role": "praktikan"}, "get_my_rentals_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": `"7"`}},
		{"tanpa klaim id menghapus key", map[string]any{"role": "praktikan"}, "get_user_profile",
			map[string]json.RawMessage{"p_caller_id": raw(`1`), "p_target_id": raw(`1`), "p_note": raw(`1`)},
			map[string]string{"p_note": "1"}},
		{"klaim id null tanpa injeksi", map[string]any{"id": nil, "role": "praktikan"}, "get_my_rentals_secure",
			map[string]json.RawMessage{"p_user_id": raw(`1`)}, map[string]string{"p_user_id": "1"}},
	}
	for _, tc := range cases {
		uc, rpc, _ := newGateway()
		if _, f := uc.Execute(context.Background(), tc.fn, sign(t, tc.claims), tc.params); f != nil {
			t.Fatalf("%s: %+v", tc.name, f)
		}
		got := map[string]string{}
		for k, v := range rpc.calls[0].params {
			got[k] = string(v)
		}
		if fmt.Sprint(got) != fmt.Sprint(tc.want) {
			t.Errorf("%s: params %v, ingin %v", tc.name, got, tc.want)
		}
	}
}

func TestGatewayDoesNotMutateClientParams(t *testing.T) {
	uc, _, _ := newGateway()
	params := map[string]json.RawMessage{"p_user_id": json.RawMessage(`1`)}
	uc.Execute(context.Background(), "get_my_rentals_secure", sign(t, map[string]any{"id": 7, "role": "praktikan"}), params)
	if string(params["p_user_id"]) != "1" {
		t.Fatal("map params milik pemanggil tidak boleh diubah")
	}
}

func TestGatewayErrors(t *testing.T) {
	token := sign(t, map[string]any{"id": 7, "role": "asisten"})
	cases := []struct {
		name   string
		err    error
		status int
		body   map[string]any
	}{
		{"pg sistem", &pgconn.PgError{Message: "permission denied for table users"}, 400, map[string]any{
			"error": "Terjadi kesalahan sistem database: permission denied for table users", "details": "permission denied for table users"}},
		{"pg bisnis", &pgconn.PgError{Message: "Akses ditolak"}, 400, map[string]any{"error": "Akses ditolak", "details": "Akses ditolak"}},
		{"fungsi tidak ada", fmt.Errorf("bungkus: %w", &repository.FunctionNotFoundError{Fn: "get_users_secure", Keys: []string{"p_x"}}), 400, map[string]any{
			"error":   "Could not find the function public.get_users_secure(p_x) in the schema cache",
			"details": "Could not find the function public.get_users_secure(p_x) in the schema cache"}},
		{"koneksi", errors.New("dial tcp: timeout"), 500, map[string]any{"error": "Internal server error"}},
	}
	for _, tc := range cases {
		uc, rpc, _ := newGateway()
		rpc.err = tc.err
		_, f := uc.Execute(context.Background(), "get_users_secure", token, nil)
		if f == nil || f.Status != tc.status || fmt.Sprint(f.Body) != fmt.Sprint(tc.body) {
			t.Errorf("%s: %+v", tc.name, f)
		}
	}
}

func TestGatewayStripsCredentials(t *testing.T) {
	uc, rpc, _ := newGateway()
	rpc.result = json.RawMessage(`[{"id":1,"password":"h","password_hash":"h","fb_password_hash":"h","fb_password_salt":"s","nama":"A"}]`)
	out, f := uc.Execute(context.Background(), "get_users_secure", sign(t, map[string]any{"id": 7, "role": "asisten"}), nil)
	if f != nil || string(out) != `[{"id":1,"nama":"A"}]` {
		t.Fatalf("out = %s, %+v", out, f)
	}
}
