package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	elhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/elearning"
	elmocks "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	elusecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

const (
	elAppURL       = "https://elearning.example.test"
	elSupabaseKey  = "rahasia-supabase-route"
	elSiakadSecret = "rahasia-siakad-route"
)

type failingLogin struct{}

func (failingLogin) Login(context.Context, string, string, string, string) (*elusecase.PraktikumLoginResult, error) {
	return nil, errors.New("akun tidak ada")
}

func newElearningRouter(t *testing.T) *gin.Engine {
	t.Helper()
	r := newTestRouter(t)
	uc := elusecase.NewAuthUsecase(&config.ElearningConfig{SupabaseJWTSecret: elSupabaseKey, SiakadJWTSecret: elSiakadSecret}, failingLogin{}, time.Now)
	users := elmocks.NewUserRepository(t)
	curriculum := elmocks.NewCurriculumRepository(t)
	ai := elusecase.AIClients{}
	RegisterElearning(r, &Elearning{
		Auth: uc,
		Handlers: ElearningHandlers{
			Auth:   elhandler.NewAuthHandler(uc, elAppURL),
			AI:     elhandler.NewAIHandler(elusecase.NewAIUsecase(uc, users, ai)),
			Grade:  elhandler.NewGradeHandler(elusecase.NewGradeUsecase(uc, users, elmocks.NewAssessmentRepository(t), ai, time.Now)),
			Report: elhandler.NewReportHandler(elusecase.NewReportUsecase(uc, curriculum, elmocks.NewProgressRepository(t), elmocks.NewReportRepository(t), time.Now)),
			Quiz:   elhandler.NewQuizHandler(elusecase.NewQuizUsecase(uc, curriculum)),
		},
	})
	return r
}

func TestElearningLegacyAPIRoutes(t *testing.T) {
	r := newElearningRouter(t)
	student := elToken(t, elSupabaseKey, jwt.MapClaims{"nim": "202711001", "nama": "Budi"})
	cases := []struct {
		name, path, body string
		header           map[string]string
		status           int
		message          string
	}{
		{"ai tanpa kunci", "/api/elearning/ai", `{"token":"x","prompt":"p"}`, nil, http.StatusInternalServerError, "AI Keys not configured on server"},
		{"grade tanpa token", "/api/elearning/grade", `{"attemptIds":["a1"]}`, nil, http.StatusUnauthorized, "Unauthorized: Missing token"},
		{"grade bearer rusak", "/api/elearning/grade", `{"token":"` + student + `"}`, map[string]string{"Authorization": "Bearer rusak"}, http.StatusUnauthorized, "Unauthorized: Invalid token"},
		{"report kosong", "/api/elearning/report", `{}`, nil, http.StatusBadRequest, "Missing token or nim"},
		{"report identitas beda", "/api/elearning/report", `{"token":"` + student + `","nim":"202711002"}`, nil, http.StatusForbidden, "Identity mismatch"},
		{"validate-quiz tanpa bearer", "/api/elearning/validate-quiz", `{"token":"` + student + `","lessonId":"l1","selectedOption":1}`, nil, http.StatusUnauthorized, "Unauthorized: Missing token"},
		{"validate-quiz field kurang", "/api/elearning/validate-quiz", `{"lessonId":"l1"}`, map[string]string{"Authorization": "Bearer " + student}, http.StatusBadRequest, "Missing required fields (lessonId, selectedOption)"},
	}
	for _, tc := range cases {
		rec := postBody(r, tc.path, "application/json", tc.body, tc.header)
		if msg, _ := envelope(t, rec); rec.Code != tc.status || msg != tc.message {
			t.Errorf("%s: %d %q, want %d %q", tc.name, rec.Code, msg, tc.status, tc.message)
		}
	}
}

func postBody(r http.Handler, path, contentType, body string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func elToken(t *testing.T, key string, claims jwt.MapClaims) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return tok
}

func envelope(t *testing.T, rec *httptest.ResponseRecorder) (message string, data map[string]any) {
	t.Helper()
	var env struct {
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body bukan JSON: %s", rec.Body.String())
	}
	return env.Message, env.Data
}

func TestElearningSSOFormToCodeExchange(t *testing.T) {
	r := newElearningRouter(t)
	siakad := elToken(t, elSiakadSecret, jwt.MapClaims{"username": "202711001", "full_name": "Budi", "role": "koordinator"})

	rec := postBody(r, "/api/elearning/sso", "application/x-www-form-urlencoded", url.Values{"token": {siakad}}.Encode(), nil)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /sso = %d, want 303", rec.Code)
	}
	loc := rec.Header().Get("Location")
	prefix := elAppURL + "/sso?code="
	if !strings.HasPrefix(loc, prefix) {
		t.Fatalf("Location = %q, want prefix %q", loc, prefix)
	}
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("Location rusak: %v", err)
	}
	code := u.Query().Get("code")
	if code == "" || strings.Contains(loc, siakad) {
		t.Fatalf("code kosong atau token bocor di URL: %q", loc)
	}

	body, _ := json.Marshal(map[string]string{"code": code})
	rec = postBody(r, "/api/elearning/receive-token", "application/json", string(body), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("tukar code = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	_, data := envelope(t, rec)
	payload, _ := data["payload"].(map[string]any)
	if payload["nim"] != "202711001" || payload["role"] != "kordas" || data["token"] == "" {
		t.Errorf("data = %v", data)
	}

	rec = postBody(r, "/api/elearning/receive-token", "application/json", string(body), nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("tukar code kedua = %d, want 401", rec.Code)
	}
}

func TestElearningSSORedirects(t *testing.T) {
	r := newElearningRouter(t)
	cases := []struct {
		name, contentType, body, want string
	}{
		{"tanpa token", "application/x-www-form-urlencoded", "", elAppURL + "/"},
		{"content-type lain", "text/plain", "token=abc", elAppURL + "/"},
		{"token rusak", "application/x-www-form-urlencoded", "token=rusak", elAppURL + "/?error=invalid_token"},
		{"field kurang", "application/json", `{"token":"` + elToken(t, elSiakadSecret, jwt.MapClaims{"username": "202711001"}) + `"}`, elAppURL + "/?error=missing_fields"},
	}
	for _, tc := range cases {
		rec := postBody(r, "/api/elearning/sso", tc.contentType, tc.body, nil)
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != tc.want {
			t.Errorf("%s: %d %q, want 302 %q", tc.name, rec.Code, rec.Header().Get("Location"), tc.want)
		}
	}

	ok := elToken(t, elSiakadSecret, jwt.MapClaims{"username": "202711001", "full_name": "Budi"})
	rec := postBody(r, "/api/elearning/sso", "application/json; charset=utf-8", `{"token":"`+ok+`"}`, nil)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("SSO JSON = %d, want 303", rec.Code)
	}
}

func TestElearningAuthEndpoints(t *testing.T) {
	r := newElearningRouter(t)

	rec := postBody(r, "/api/elearning/verify", "application/json", `{"token":"rusak"}`, nil)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusUnauthorized || msg != "Invalid token: Invalid Compact JWS" {
		t.Errorf("verify rusak = %d %q", rec.Code, msg)
	}

	rec = postBody(r, "/api/elearning/verify", "application/json", "", nil)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusBadRequest || msg != "Token is required" {
		t.Errorf("verify tanpa body = %d %q", rec.Code, msg)
	}

	rec = postBody(r, "/api/elearning/receive-token", "application/json", `{}`, nil)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusBadRequest || msg != `Field "token" wajib disertakan di request body.` {
		t.Errorf("receive-token kosong = %d %q", rec.Code, msg)
	}

	rec = postBody(r, "/api/elearning/login", "application/json", `{"identifier":"202711001","password":"salah"}`, nil)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusUnauthorized || msg != "NIM/email atau password salah." {
		t.Errorf("login gagal = %d %q", rec.Code, msg)
	}

	rec = postBody(r, "/api/elearning/auth/supabase-token", "", "", nil)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusUnauthorized || msg != "Unauthorized: Missing token" {
		t.Errorf("supabase-token tanpa Bearer = %d %q", rec.Code, msg)
	}
	rec = postBody(r, "/api/elearning/auth/supabase-token", "", "", map[string]string{"Authorization": "Bearer rusak"})
	if msg, _ := envelope(t, rec); rec.Code != http.StatusUnauthorized || msg != "Unauthorized: Invalid token" {
		t.Errorf("supabase-token token rusak = %d %q", rec.Code, msg)
	}

	bearer := elToken(t, elSupabaseKey, jwt.MapClaims{"nim": "202711001", "nama": "Budi", "user_role": "asisten", "role": "authenticated"})
	rec = postBody(r, "/api/elearning/auth/supabase-token", "", "", map[string]string{"Authorization": "Bearer " + bearer})
	if _, data := envelope(t, rec); rec.Code != http.StatusOK || data["token"] == "" {
		t.Errorf("supabase-token = %d %v", rec.Code, data)
	}
}

func TestElearningLoginRateLimit(t *testing.T) {
	r := newElearningRouter(t)
	ip := map[string]string{"X-Forwarded-For": "203.0.113.9"}
	for i := range 10 {
		if rec := postBody(r, "/api/elearning/login", "application/json", `{}`, ip); rec.Code != http.StatusBadRequest {
			t.Fatalf("login ke-%d = %d, want 400", i+1, rec.Code)
		}
	}
	if rec := postBody(r, "/api/elearning/login", "application/json", `{}`, ip); rec.Code != http.StatusTooManyRequests {
		t.Errorf("login ke-11 = %d, want 429", rec.Code)
	}
}

func TestElearningUnavailableWithoutDB(t *testing.T) {
	r := newTestRouter(t)
	RegisterElearningUnavailable(r)
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/elearning/sso"},
		{http.MethodGet, "/api/elearning/curriculum/levels"},
		{http.MethodPost, "/api/elearning/login"},
	} {
		rec := serve(r, tc.method, tc.path, "", nil)
		if msg, _ := envelope(t, rec); rec.Code != http.StatusServiceUnavailable || msg != "Backend elearning belum dikonfigurasi (DATABASE_URL_ELEARNING)." {
			t.Errorf("%s %s = %d %q, want 503", tc.method, tc.path, rec.Code, msg)
		}
	}
	if rec := serve(r, http.MethodGet, "/api/health", "", nil); rec.Code != http.StatusOK {
		t.Errorf("health = %d, want 200", rec.Code)
	}
}
