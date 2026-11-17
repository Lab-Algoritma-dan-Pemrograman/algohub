package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

type fakeSiakadGemini struct{ calls int }

func (f *fakeSiakadGemini) GenerateContent(context.Context, string, string, string, string) (string, error) {
	f.calls++
	return `[{"day_of_week":"Senin","start_time":"06:00","end_time":"07:00"}]`, nil
}

type fakeSiakadSSO struct{}

func (fakeSiakadSSO) HandshakeUserID(_ context.Context, code string, _ time.Time) (userID *int64, found bool, err error) {
	id := int64(9)
	return &id, code == "valid", nil
}

func (fakeSiakadSSO) User(context.Context, int64) (entity.SSOUser, bool, error) {
	return entity.SSOUser{NIM: "202711001", FullName: "Budi", Role: "kordas"}, true, nil
}

func (fakeSiakadSSO) DeleteHandshake(context.Context, string) error { return nil }

func (fakeSiakadSSO) StudentSchedule(context.Context, int64) (entity.ScheduleClass, error) {
	return entity.ScheduleClass{}, nil
}

func (fakeSiakadSSO) AssistantSchedule(context.Context, int64) (entity.ScheduleClass, error) {
	return entity.ScheduleClass{}, nil
}

type fakeSiakadElearning struct{}

func (fakeSiakadElearning) Levels(context.Context) ([]entity.ElearningLevel, error) { return nil, nil }

func (fakeSiakadElearning) StudentProgress(context.Context, string) ([]entity.ElearningStudentProgress, error) {
	return nil, nil
}

func (fakeSiakadElearning) ActiveSession(context.Context, string) (*entity.ElearningSession, error) {
	return nil, nil
}

func (fakeSiakadElearning) Modules(context.Context) ([]entity.ElearningModule, error) {
	return nil, nil
}

func (fakeSiakadElearning) Lessons(context.Context) ([]entity.ElearningLesson, error) {
	return nil, nil
}

func (fakeSiakadElearning) LevelOverrides(context.Context, string) (json.RawMessage, error) {
	return nil, nil
}

type fakeSiakadProgress struct{ upserts int }

func (f *fakeSiakadProgress) Upsert(context.Context, *entity.ElearningProgressRow) error {
	f.upserts++
	return nil
}

func (f *fakeSiakadProgress) ExistsByNIM(context.Context, string) (bool, error) { return false, nil }

func (f *fakeSiakadProgress) Update(context.Context, *entity.ElearningProgressRow) error { return nil }

func (f *fakeSiakadProgress) Insert(context.Context, *entity.ElearningProgressRow) error { return nil }

func newIntegrationRouter(t *testing.T) (*gin.Engine, *fakeSiakadGemini, *fakeSiakadProgress) {
	t.Helper()
	tokens := usecase.NewTokens(siakadSecret)
	g := &fakeSiakadGemini{}
	progress := &fakeSiakadProgress{}
	r := newTestRouter(t)
	RegisterSiakad(r, &Siakad{Handlers: SiakadHandlers{
		Integration: handler.NewIntegrationHandler(
			usecase.NewScheduleAIUsecase(tokens, usecase.AIProviders{Gemini: g}),
			usecase.NewSSOUsecase(fakeSiakadSSO{}, tokens),
			usecase.NewElearningSyncUsecase(tokens, fakeSiakadElearning{}, progress),
		),
	}})
	return r, g, progress
}

func TestSiakadAnalyzeSchedule(t *testing.T) {
	r, g, _ := newIntegrationRouter(t)
	body := `{"fileBase64":"QUJD","mimeType":"image/png"}`
	staff := siakadToken(t, map[string]any{"role": "asisten"})

	rec := send(r, http.MethodPost, "/api/siakad/analyze-schedule", body, map[string]string{"Authorization": "Bearer " + staff})
	if rec.Code != http.StatusOK || rec.Body.String() != `{"slots":[{"day_of_week":"Senin","start_time":"06:00","end_time":"07:00"}]}` {
		t.Fatalf("bearer staf: %d %s", rec.Code, rec.Body)
	}
	rec = send(r, http.MethodPost, "/api/siakad/analyze-schedule", body, map[string]string{"Cookie": "siakad_token=" + staff})
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Missing session token"}` {
		t.Fatalf("cookie saja (T-15): %d %s", rec.Code, rec.Body)
	}
	rec = send(r, http.MethodPost, "/api/siakad/analyze-schedule", body, map[string]string{"Authorization": "Bearer "})
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Invalid or expired token"}` {
		t.Fatalf("bearer kosong: %d %s", rec.Code, rec.Body)
	}
	rec = send(r, http.MethodPost, "/api/siakad/analyze-schedule", `{"fileBase64":1}`, map[string]string{"Authorization": "Bearer " + staff})
	if rec.Code != http.StatusBadRequest || rec.Body.String() != `{"error":"fileBase64 and mimeType are required"}` {
		t.Fatalf("body salah tipe: %d %s", rec.Code, rec.Body)
	}
	if g.calls != 1 {
		t.Fatalf("gemini dipanggil %d kali", g.calls)
	}
}

func TestSiakadGenerateJWT(t *testing.T) {
	r, _, _ := newIntegrationRouter(t)
	rec := send(r, http.MethodPost, "/api/siakad/generate-jwt", `{}`, nil)
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Missing handshake code"}` {
		t.Fatalf("tanpa kode: %d %s", rec.Code, rec.Body)
	}
	rec = send(r, http.MethodPost, "/api/siakad/generate-jwt", `{"handshake_code":"lama"}`, nil)
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Sesi verifikasi kadaluarsa. Harap coba lagi."}` {
		t.Fatalf("kadaluarsa: %d %s", rec.Code, rec.Body)
	}
	rec = send(r, http.MethodPost, "/api/siakad/generate-jwt", `{"handshake_code":"valid"}`, nil)
	var out struct{ Token string }
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &out) != nil {
		t.Fatalf("valid: %d %s", rec.Code, rec.Body)
	}
	claims, err := usecase.NewTokens(siakadSecret).Verify(out.Token)
	if err != nil || claims["nim"] != "202711001" || claims["nama"] != "Budi" || claims["role"] != "kordas" {
		t.Fatalf("klaim %v %v", claims, err)
	}
}

func TestSiakadSyncElearning(t *testing.T) {
	r, _, progress := newIntegrationRouter(t)
	token := siakadToken(t, map[string]any{"nim": "202711001"})
	rec := send(r, http.MethodPost, "/api/siakad/sync-elearning", "", map[string]string{"Authorization": "Bearer " + token})
	want := `{"nim":"202711001","student_name":"Mahasiswa","completed_lessons":0,"total_lessons":0,"completion_percentage":0,"is_completed":false,"completed_levels":[],"current_level":"Belum Mulai","last_accessed_at":null}`
	if rec.Code != http.StatusOK || rec.Body.String() != want || progress.upserts != 1 {
		t.Fatalf("%d %s, upsert %d", rec.Code, rec.Body, progress.upserts)
	}
	rec = send(r, http.MethodPost, "/api/siakad/sync-elearning", "", map[string]string{"Cookie": "siakad_token=" + token})
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Missing session token"}` {
		t.Fatalf("cookie saja (T-15): %d %s", rec.Code, rec.Body)
	}
}

func TestSiakadIntegrationMethods(t *testing.T) {
	r, _, _ := newIntegrationRouter(t)
	for _, path := range []string{"/api/siakad/analyze-schedule", "/api/siakad/generate-jwt", "/api/siakad/sync-elearning"} {
		for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			if rec := send(r, m, path, "", nil); rec.Code != http.StatusMethodNotAllowed || rec.Body.String() != `{"error":"Method not allowed"}` {
				t.Errorf("%s %s: %d %s", m, path, rec.Code, rec.Body)
			}
		}
	}
}
