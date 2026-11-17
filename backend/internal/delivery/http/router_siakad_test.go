package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

const siakadSecret = "rahasia-uji"

type fakeSiakadAuth struct {
	row     json.RawMessage
	revoked map[string]bool
	revokes []string
}

func (f *fakeSiakadAuth) LoginUser(context.Context, string, string) (json.RawMessage, error) {
	return f.row, nil
}

func (f *fakeSiakadAuth) ActiveShift(context.Context) (string, bool) { return "all", true }

func (f *fakeSiakadAuth) IsTokenRevoked(_ context.Context, jti string) (bool, error) {
	return f.revoked[jti], nil
}

func (f *fakeSiakadAuth) RevokeToken(_ context.Context, jti, _ string, _ time.Time) error {
	f.revokes = append(f.revokes, jti)
	return nil
}

type fakeSiakadRPC struct {
	calls []string
}

func (f *fakeSiakadRPC) Call(_ context.Context, fn string, _ map[string]json.RawMessage) (json.RawMessage, error) {
	f.calls = append(f.calls, fn)
	return json.RawMessage(`[{"id":1, "nama":"A"}]`), nil
}

func newSiakadRouter(t *testing.T) (*gin.Engine, *fakeSiakadAuth, *fakeSiakadRPC) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("87654321"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	row, _ := json.Marshal(map[string]any{
		"id": 7, "username": "202711001", "role": "asisten", "password": string(hash), "is_active": true, "shift": nil,
	})
	auth := &fakeSiakadAuth{row: row, revoked: map[string]bool{}}
	rpc := &fakeSiakadRPC{}
	tokens := usecase.NewTokens(siakadSecret)

	r := newTestRouter(t)
	RegisterSiakad(r, &Siakad{Handlers: SiakadHandlers{
		Auth: handler.NewAuthHandler(usecase.NewAuthUsecase(auth, tokens)),
		RPC:  handler.NewRPCHandler(usecase.NewRPCUsecase(rpc, auth, tokens)),
	}})
	return r, auth, rpc
}

func send(r http.Handler, method, path, body string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func siakadToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	token, err := usecase.NewTokens(siakadSecret).Sign(claims, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestSiakadLoginSetsCookie(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	rec := send(r, http.MethodPost, "/api/siakad/auth/login", `{"username":"202711001","password":"87654321"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var body struct {
		User  map[string]any `json:"user"`
		Token string         `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Token == "" || body.User["username"] != "202711001" {
		t.Fatalf("body %s", rec.Body)
	}
	if _, ok := body.User["password"]; ok {
		t.Fatal("hash password bocor")
	}
	want := "siakad_token=" + body.Token + "; Domain=.algohub.web.id; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Lax"
	if got := rec.Header().Values("Set-Cookie"); len(got) != 1 || got[0] != want {
		t.Fatalf("Set-Cookie = %q\ningin %q", got, want)
	}
}

func TestSiakadLoginRateLimit(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	ip := map[string]string{"X-Forwarded-For": "203.0.113.50"}
	for i := range 5 {
		if rec := send(r, http.MethodPost, "/api/siakad/auth/login", `{}`, ip); rec.Code != http.StatusBadRequest {
			t.Fatalf("request ke-%d = %d, ingin 400", i+1, rec.Code)
		}
	}
	rec := send(r, http.MethodPost, "/api/siakad/auth/login", `{"username":"202711001","password":"87654321"}`, ip)
	if rec.Code != http.StatusTooManyRequests || rec.Body.String() != `{"error":"Terlalu banyak percobaan login. Silakan coba lagi dalam satu menit."}` {
		t.Fatalf("request ke-6 = %d %s", rec.Code, rec.Body)
	}
	if rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=register_user", `{}`, ip); rec.Code != http.StatusOK {
		t.Fatalf("kuota rpc harus terpisah dari login, dapat %d", rec.Code)
	}
}

func TestSiakadLoginBadBodyIsEmpty(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	for _, body := range []string{"", "bukan json", `{"username":5,"password":"x"}`, `[]`} {
		rec := send(r, http.MethodPost, "/api/siakad/auth/login", body, map[string]string{"X-Forwarded-For": "203.0.113.60"})
		if rec.Code != http.StatusBadRequest || rec.Body.String() != `{"error":"Username dan password wajib diisi."}` {
			t.Errorf("body %q: %d %s", body, rec.Code, rec.Body)
		}
	}
}

func TestSiakadLogout(t *testing.T) {
	r, auth, _ := newSiakadRouter(t)
	token := siakadToken(t, map[string]any{"id": 7, "jti": "jti-out"})
	rec := send(r, http.MethodPost, "/api/siakad/auth/logout", "", map[string]string{"Cookie": "siakad_token=" + token})
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("status %d body %q", rec.Code, rec.Body)
	}
	want := "siakad_token=; Domain=.algohub.web.id; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax"
	if got := rec.Header().Get("Set-Cookie"); got != want {
		t.Fatalf("Set-Cookie = %q", got)
	}
	if len(auth.revokes) != 1 || auth.revokes[0] != "jti-out" {
		t.Fatalf("revoke_token = %v", auth.revokes)
	}

	rec = send(r, http.MethodPost, "/api/siakad/auth/logout", "", nil)
	if rec.Code != http.StatusNoContent || rec.Header().Get("Set-Cookie") != want {
		t.Fatalf("logout tanpa token: %d %q", rec.Code, rec.Header().Get("Set-Cookie"))
	}
}

func TestSiakadVerify(t *testing.T) {
	r, auth, _ := newSiakadRouter(t)
	auth.revoked["dicabut"] = true
	token := siakadToken(t, map[string]any{"id": 7, "role": "asisten", "jti": "aktif"})

	cases := []struct {
		name   string
		method string
		header map[string]string
	}{
		{"cookie GET", http.MethodGet, map[string]string{"Cookie": "siakad_token=" + token}},
		{"header POST", http.MethodPost, map[string]string{"Authorization": "Bearer " + token}},
		{"cookie menang atas header", http.MethodPost, map[string]string{"Cookie": "siakad_token=" + token, "Authorization": "Bearer rusak"}},
	}
	for _, tc := range cases {
		rec := send(r, tc.method, "/api/siakad/auth/verify", "", tc.header)
		if rec.Code != http.StatusOK || rec.Body.String() != `{"user":{"id":7,"role":"asisten"}}` {
			t.Errorf("%s: %d %s", tc.name, rec.Code, rec.Body)
		}
		if rec.Header().Get("Set-Cookie") != "" {
			t.Errorf("%s: cookie tidak boleh diubah", tc.name)
		}
	}

	revoked := siakadToken(t, map[string]any{"id": 7, "role": "asisten", "jti": "dicabut"})
	rec := send(r, http.MethodPost, "/api/siakad/auth/verify", "", map[string]string{"Authorization": "Bearer " + revoked})
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Session telah berakhir. Silakan login kembali."}` {
		t.Fatalf("dicabut: %d %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatal("cookie harus dihapus untuk token dicabut")
	}

	rec = send(r, http.MethodGet, "/api/siakad/auth/verify", "", nil)
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Missing token"}` || rec.Header().Get("Set-Cookie") != "" {
		t.Fatalf("tanpa token: %d %s", rec.Code, rec.Body)
	}
}

func TestSiakadMethodNotAllowed(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	cases := []struct{ method, path string }{
		{http.MethodGet, "/api/siakad/rpc"},
		{http.MethodGet, "/api/siakad/auth/login"},
		{http.MethodDelete, "/api/siakad/auth/logout"},
		{http.MethodPut, "/api/siakad/auth/verify"},
		{http.MethodOptions, "/api/siakad/rpc"},
	}
	for _, tc := range cases {
		rec := send(r, tc.method, tc.path, "", nil)
		if rec.Code != http.StatusMethodNotAllowed || rec.Body.String() != `{"error":"Method not allowed"}` {
			t.Errorf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body)
		}
	}
}

func TestSiakadRPC(t *testing.T) {
	r, _, rpc := newSiakadRouter(t)
	bearer := map[string]string{"Authorization": "Bearer " + siakadToken(t, map[string]any{"id": 7, "role": "asisten"})}

	rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=get_users_secure", `{"p_x":1}`, bearer)
	if rec.Code != http.StatusOK || rec.Body.String() != `[{"id":1, "nama":"A"}]` {
		t.Fatalf("sukses: %d %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}

	cases := []struct {
		name, path, body string
		status           int
		want             string
	}{
		{"tanpa fn", "/api/siakad/rpc", `{}`, 400, `{"error":"Missing function name"}`},
		{"body array", "/api/siakad/rpc?fn=get_users_secure", `[1]`, 400, `{"error":"Invalid JSON body"}`},
		{"body rusak", "/api/siakad/rpc?fn=get_users_secure", `{`, 400, `{"error":"Invalid JSON body"}`},
		{"tanpa token", "/api/siakad/rpc?fn=get_users_secure", `{}`, 401, `{"error":"Unauthorized: Missing session token"}`},
	}
	for _, tc := range cases {
		h := bearer
		if tc.status == http.StatusUnauthorized {
			h = nil
		}
		rec := send(r, http.MethodPost, tc.path, tc.body, h)
		if rec.Code != tc.status || rec.Body.String() != tc.want {
			t.Errorf("%s: %d %s", tc.name, rec.Code, rec.Body)
		}
	}
	if rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=get_public_settings", "", nil); rec.Code != http.StatusOK {
		t.Fatalf("RPC publik dengan body kosong: %d %s", rec.Code, rec.Body)
	}
	if len(rpc.calls) != 2 {
		t.Fatalf("calls = %v", rpc.calls)
	}
}

func TestSiakadRPCRateLimit(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	ip := map[string]string{"X-Forwarded-For": "203.0.113.70"}
	for range 7 {
		if rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=get_public_settings", "", ip); rec.Code != http.StatusOK {
			t.Fatalf("RPC selain login_user/register_user tidak dibatasi, dapat %d", rec.Code)
		}
	}
	for i := range 5 {
		if rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=register_user", "", ip); rec.Code != http.StatusOK {
			t.Fatalf("register_user ke-%d = %d", i+1, rec.Code)
		}
	}
	rec := send(r, http.MethodPost, "/api/siakad/rpc?fn=login_user", "", ip)
	if rec.Code != http.StatusTooManyRequests || rec.Body.String() != `{"error":"Terlalu banyak percobaan. Silakan coba lagi dalam satu menit."}` {
		t.Fatalf("ke-6 = %d %s", rec.Code, rec.Body)
	}
}

func TestSiakadCORS(t *testing.T) {
	r, _, _ := newSiakadRouter(t)
	rec := serve(r, http.MethodOptions, "/api/siakad/rpc", allowedOrigin, map[string]string{
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "authorization, content-type",
	})
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != allowedOrigin {
		t.Fatalf("preflight: %d %v", rec.Code, rec.Header())
	}
	rec = serve(r, http.MethodPost, "/api/siakad/rpc?fn=get_users_secure", allowedOrigin, nil)
	if rec.Code != http.StatusUnauthorized || rec.Header().Get("Access-Control-Allow-Origin") != allowedOrigin ||
		rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("401 tanpa CORS: %d %v", rec.Code, rec.Header())
	}
}
