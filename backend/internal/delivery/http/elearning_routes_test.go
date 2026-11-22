package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	elhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
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

type elRepos struct {
	users        *elmocks.UserRepository
	curriculum   *elmocks.CurriculumRepository
	progress     *elmocks.ProgressRepository
	achievements *elmocks.AchievementRepository
}

func newElearningRouter(t *testing.T) *gin.Engine {
	t.Helper()
	r, _ := newElearningRouterRepos(t)
	return r
}

func newElearningRouterRepos(t *testing.T) (*gin.Engine, *elRepos) {
	t.Helper()
	r := newTestRouter(t)
	uc := elusecase.NewAuthUsecase(&config.ElearningConfig{SupabaseJWTSecret: elSupabaseKey, SiakadJWTSecret: elSiakadSecret}, failingLogin{}, time.Now)
	m := &elRepos{
		users:        elmocks.NewUserRepository(t),
		curriculum:   elmocks.NewCurriculumRepository(t),
		progress:     elmocks.NewProgressRepository(t),
		achievements: elmocks.NewAchievementRepository(t),
	}
	ai := elusecase.AIClients{}
	logs := elusecase.NewActivityLogUsecase(elmocks.NewActivityLogRepository(t), time.Now)
	achievements := elusecase.NewAchievementUsecase(m.achievements, m.users, m.progress, time.Now)
	learning := elusecase.NewProgressUsecase(m.users, m.progress, m.curriculum, achievements, time.Now)
	RegisterElearning(r, &Elearning{
		Auth: uc,
		Handlers: ElearningHandlers{
			Auth:        elhandler.NewAuthHandler(uc, elAppURL),
			AI:          elhandler.NewAIHandler(elusecase.NewAIUsecase(uc, m.users, ai)),
			Grade:       elhandler.NewGradeHandler(elusecase.NewGradeUsecase(uc, m.users, elmocks.NewAssessmentRepository(t), ai, time.Now)),
			Report:      elhandler.NewReportHandler(elusecase.NewReportUsecase(uc, m.curriculum, m.progress, elmocks.NewReportRepository(t), time.Now)),
			Quiz:        elhandler.NewQuizHandler(elusecase.NewQuizUsecase(uc, m.curriculum)),
			Curriculum:  elhandler.NewCurriculumHandler(elusecase.NewCurriculumUsecase(m.curriculum)),
			Progress:    elhandler.NewProgressHandler(learning),
			Profile:     elhandler.NewProfileHandler(elusecase.NewProfileUsecase(m.users, achievements, time.Now), logs),
			Achievement: elhandler.NewAchievementHandler(achievements),
			AdminUser:   elhandler.NewAdminUserHandler(elusecase.NewAdminUserUsecase(m.users, m.progress, m.curriculum, m.achievements, achievements, logs)),
			Game:        elhandler.NewGameHandler(elusecase.NewGameUsecase(elmocks.NewGameRepository(t), m.users, time.Now, nil)),
			Leaderboard: elhandler.NewLeaderboardHandler(elusecase.NewLeaderboardUsecase(elmocks.NewLeaderboardRepository(t), time.Now)),
			Prestige:    elhandler.NewPrestigeHandler(elusecase.NewPrestigeUsecase(m.users, m.curriculum, m.progress, elmocks.NewPrestigeRepository(t), logs, time.Now)),
			Playground:  elhandler.NewPlaygroundHandler(elusecase.NewPlaygroundUsecase(elmocks.NewPlaygroundRepository(t))),
			Monitoring:  elhandler.NewMonitoringHandler(elusecase.NewSessionUsecase(elmocks.NewSessionRepository(t), m.users, m.progress, learning, time.Now), logs),
			Storage:     elhandler.NewStorageHandler(elusecase.NewStorageUsecase(rejectUploader{}, time.Now)),
		},
	})
	return r, m
}

type rejectUploader struct{}

func (rejectUploader) Upload(string, []byte, string, bool) (string, error) {
	return "", errors.New("upload tidak boleh dipanggil di test route")
}

var learningRoutes = []struct{ method, path string }{
	{http.MethodGet, "/api/elearning/curriculum/levels"},
	{http.MethodGet, "/api/elearning/curriculum/modules"},
	{http.MethodGet, "/api/elearning/curriculum/lessons"},
	{http.MethodPut, "/api/elearning/curriculum"},
	{http.MethodDelete, "/api/elearning/curriculum"},
	{http.MethodGet, "/api/elearning/progress/me/lessons"},
	{http.MethodGet, "/api/elearning/progress/me/recent"},
	{http.MethodPost, "/api/elearning/progress/lessons/l1/complete"},
	{http.MethodPost, "/api/elearning/progress/xp"},
	{http.MethodPost, "/api/elearning/progress/study-time"},
	{http.MethodPost, "/api/elearning/streak"},
	{http.MethodPost, "/api/elearning/profile/sync"},
	{http.MethodGet, "/api/elearning/achievements"},
	{http.MethodGet, "/api/elearning/achievements/me"},
	{http.MethodPost, "/api/elearning/achievements/check"},
	{http.MethodPost, "/api/elearning/achievements/check-xp"},
	{http.MethodGet, "/api/elearning/admin/users"},
	{http.MethodGet, "/api/elearning/admin/users/202711002/progress"},
	{http.MethodPatch, "/api/elearning/admin/users/202711002/level-access"},
	{http.MethodPatch, "/api/elearning/admin/users/202711002/role"},
	{http.MethodPost, "/api/elearning/admin/users/202711002/reset"},
	{http.MethodPost, "/api/elearning/admin/users/202711002/reset-level"},
	{http.MethodPost, "/api/elearning/admin/users/202711002/reset-lesson"},
	{http.MethodPut, "/api/elearning/admin/users/202711002/xp"},
	{http.MethodDelete, "/api/elearning/admin/users/202711002"},
	{http.MethodPost, "/api/elearning/activity-logs"},
	{http.MethodGet, "/api/elearning/game/questions"},
	{http.MethodPost, "/api/elearning/game/questions"},
	{http.MethodPut, "/api/elearning/game/questions"},
	{http.MethodPost, "/api/elearning/game/questions/seed"},
	{http.MethodPatch, "/api/elearning/game/questions/q1"},
	{http.MethodDelete, "/api/elearning/game/questions/q1"},
	{http.MethodGet, "/api/elearning/game/settings"},
	{http.MethodPut, "/api/elearning/game/settings"},
	{http.MethodGet, "/api/elearning/game/plays/me"},
	{http.MethodPost, "/api/elearning/game/history"},
	{http.MethodGet, "/api/elearning/leaderboard"},
	{http.MethodGet, "/api/elearning/leaderboard/rank"},
	{http.MethodGet, "/api/elearning/leaderboard/filters"},
	{http.MethodGet, "/api/elearning/leaderboard/weekly"},
	{http.MethodGet, "/api/elearning/prestige/eligibility"},
	{http.MethodPost, "/api/elearning/prestige/reset"},
	{http.MethodGet, "/api/elearning/playground/examples"},
	{http.MethodPost, "/api/elearning/playground/examples"},
	{http.MethodPatch, "/api/elearning/playground/examples/e1"},
	{http.MethodDelete, "/api/elearning/playground/examples/e1"},
	{http.MethodPut, "/api/elearning/sessions/me"},
	{http.MethodDelete, "/api/elearning/sessions/me"},
	{http.MethodGet, "/api/elearning/monitoring/sessions"},
	{http.MethodDelete, "/api/elearning/monitoring/sessions/stale"},
	{http.MethodGet, "/api/elearning/monitoring/students"},
	{http.MethodGet, "/api/elearning/monitoring/progress"},
	{http.MethodGet, "/api/elearning/activity-logs"},
	{http.MethodDelete, "/api/elearning/activity-logs/log-1"},
	{http.MethodPost, "/api/elearning/storage/lesson-images"},
}

func TestElearningLearningRoutesRequireToken(t *testing.T) {
	r := newElearningRouter(t)
	for _, rt := range learningRoutes {
		rec := serve(r, rt.method, rt.path, "", nil)
		msg, _ := envelope(t, rec)
		if rec.Code != http.StatusUnauthorized || msg != "Unauthorized: Missing token" {
			t.Errorf("%s %s = %d %q, want 401", rt.method, rt.path, rec.Code, msg)
		}
		if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
			t.Errorf("%s %s Cache-Control = %q", rt.method, rt.path, got)
		}
	}
}

func TestElearningLearningRoutesWithToken(t *testing.T) {
	r, m := newElearningRouterRepos(t)
	bearer := map[string]string{"Authorization": "Bearer " + elToken(t, elSupabaseKey, jwt.MapClaims{"nim": "202711001", "nama": "Budi", "user_role": "praktikan"})}

	m.curriculum.On("ListLevels", mock.Anything, 2, 100).Return([]entity.Level{{ID: "level-1"}}, int64(101), nil)
	rec := serve(r, http.MethodGet, "/api/elearning/curriculum/levels?page=2&limit=500", "", bearer)
	_, data := envelope(t, rec)
	if rec.Code != http.StatusOK || data["total"] != float64(101) || data["limit"] != float64(100) || rec.Header().Get("Cache-Control") != "private, no-store" {
		t.Errorf("levels = %d %v %q", rec.Code, data, rec.Header().Get("Cache-Control"))
	}

	rec = sendJSON(r, http.MethodPut, "/api/elearning/curriculum", `{"levels":[{"id":"a"}]}`, bearer)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusForbidden || msg != `new row violates row-level security policy for table "levels"` {
		t.Errorf("PUT curriculum praktikan = %d %q", rec.Code, msg)
	}

	rec = sendJSON(r, http.MethodPut, "/api/elearning/curriculum", `{bukan json`, bearer)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusBadRequest || msg != "Input tidak valid" {
		t.Errorf("PUT curriculum body rusak = %d %q", rec.Code, msg)
	}
}

func TestElearningGameMonitoringStorageRoutesWithToken(t *testing.T) {
	r := newElearningRouter(t)
	bearer := map[string]string{"Authorization": "Bearer " + elToken(t, elSupabaseKey, jwt.MapClaims{"nim": "202711001", "nama": "Budi", "user_role": "praktikan"})}

	rec := sendJSON(r, http.MethodPut, "/api/elearning/game/settings", `{"bugHuntActive":false}`, bearer)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusForbidden || msg != `new row violates row-level security policy for table "game_settings"` {
		t.Errorf("PUT game/settings praktikan = %d %q", rec.Code, msg)
	}

	rec = serve(r, http.MethodGet, "/api/elearning/monitoring/sessions", "", bearer)
	if _, data := envelope(t, rec); rec.Code != http.StatusOK || data["total"] != float64(0) || len(data["items"].([]any)) != 0 {
		t.Errorf("GET monitoring/sessions praktikan = %d %v", rec.Code, data)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="a.png"`)
	h.Set("Content-Type", "image/png")
	part, _ := mw.CreatePart(h)
	_, _ = part.Write([]byte("png"))
	_ = mw.Close()
	rec = postBody(r, "/api/elearning/storage/lesson-images", mw.FormDataContentType(), body.String(), bearer)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusForbidden || msg != "new row violates row-level security policy" {
		t.Errorf("upload praktikan = %d %q", rec.Code, msg)
	}

	rec = postBody(r, "/api/elearning/storage/lesson-images", "application/json", `{}`, bearer)
	if msg, _ := envelope(t, rec); rec.Code != http.StatusBadRequest || msg != "Input tidak valid" {
		t.Errorf("upload tanpa file = %d %q", rec.Code, msg)
	}
}

func sendJSON(r http.Handler, method, path, body string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
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
	ip := map[string]string{"X-Vercel-Forwarded-For": "203.0.113.9"}
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
