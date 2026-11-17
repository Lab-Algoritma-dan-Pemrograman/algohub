package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"testing"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/bynara"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
)

type aiCall struct {
	model, prompt, mimeType, data string
}

type fakeGemini struct {
	calls []aiCall
	text  string
	err   error
}

func (f *fakeGemini) GenerateContent(_ context.Context, model, prompt, mimeType, data string) (string, error) {
	f.calls = append(f.calls, aiCall{model, prompt, mimeType, data})
	return f.text, f.err
}

type fakeBynara struct {
	calls []aiCall
	text  string
	err   error
}

func (f *fakeBynara) ChatImage(_ context.Context, model, prompt, dataURL string) (string, error) {
	f.calls = append(f.calls, aiCall{model: model, prompt: prompt, data: dataURL})
	return f.text, f.err
}

const slotsJSON = `[{"day_of_week":"Senin","start_time":"06:00","end_time":"12:00"}]`

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := encodeJSON(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func staffBearer(t *testing.T) string {
	return sign(t, map[string]any{"id": 7, "role": "Asisten"})
}

func TestSchedulePromptMatchesLegacy(t *testing.T) {
	// sha256 template literal analyze-schedule.ts baris 76–93 dengan akhir baris LF.
	sum := sha256.Sum256([]byte(schedulePrompt))
	if got := hex.EncodeToString(sum[:]); len(schedulePrompt) != 986 || got != "9a5116a018374d77971a37a524b375728d276866797ac97dc76d298036c724e5" {
		t.Fatalf("prompt berubah: %d byte, sha256 %s", len(schedulePrompt), got)
	}
}

func TestAnalyzeScheduleGemini(t *testing.T) {
	cases := []struct {
		model, want string
	}{
		{"", "gemini-3-flash"},
		{"gemini-2.5-flash", "gemini-2.5-flash"},
		{"gemini-9", "gemini-3-flash"},
	}
	for _, tc := range cases {
		g := &fakeGemini{text: "```JSON\n" + slotsJSON + "\n```"}
		b := &fakeBynara{}
		uc := NewScheduleAIUsecase(NewTokens(testSecret), AIProviders{Gemini: g, Bynara: b})
		body, f := uc.Analyze(context.Background(), staffBearer(t), true, dto.AnalyzeScheduleRequest{FileBase64: "QUJD", MimeType: "application/pdf", Model: tc.model})
		if f != nil {
			t.Fatalf("%s: failure %+v", tc.model, f)
		}
		if got := string(mustJSON(t, body)); got != `{"slots":`+slotsJSON+`}` {
			t.Errorf("%s: body %s", tc.model, got)
		}
		if len(g.calls) != 1 || len(b.calls) != 0 {
			t.Fatalf("%s: gemini %d bynara %d", tc.model, len(g.calls), len(b.calls))
		}
		if c := g.calls[0]; c != (aiCall{tc.want, schedulePrompt, "application/pdf", "QUJD"}) {
			t.Errorf("%s: panggilan %+v", tc.model, c)
		}
	}
}

func TestAnalyzeScheduleBynara(t *testing.T) {
	g := &fakeGemini{}
	b := &fakeBynara{text: "Berikut jadwalnya: " + slotsJSON + " semoga membantu"}
	uc := NewScheduleAIUsecase(NewTokens(testSecret), AIProviders{Gemini: g, Bynara: b})
	body, f := uc.Analyze(context.Background(), sign(t, map[string]any{"role": "koordinator"}), true,
		dto.AnalyzeScheduleRequest{FileBase64: "QUJD", MimeType: "image/png", Model: "stepfun-3.7-flash"})
	if f != nil || string(mustJSON(t, body)) != `{"slots":`+slotsJSON+`}` {
		t.Fatalf("body %v, failure %+v", body, f)
	}
	if len(g.calls) != 0 || len(b.calls) != 1 || b.calls[0] != (aiCall{model: "stepfun-3.7-flash", prompt: schedulePrompt, data: "data:image/png;base64,QUJD"}) {
		t.Fatalf("panggilan gemini %+v bynara %+v", g.calls, b.calls)
	}
}

func TestAnalyzeScheduleFailures(t *testing.T) {
	long := strings.Repeat("é", 400)
	bearer := staffBearer(t)
	img := dto.AnalyzeScheduleRequest{FileBase64: "QUJD", MimeType: "image/png"}
	stepfun := img
	stepfun.Model = "stepfun-3.7-flash"
	pdfStepfun := stepfun
	pdfStepfun.MimeType = "application/pdf"

	cases := []struct {
		name      string
		token     string
		hasBearer bool
		ai        AIProviders
		req       dto.AnalyzeScheduleRequest
		status    int
		body      string
	}{
		{"tanpa bearer", "", false, AIProviders{}, img, 401, `{"error":"Unauthorized: Missing session token"}`},
		{"token kosong", "", true, AIProviders{}, img, 401, `{"error":"Unauthorized: Invalid or expired token"}`},
		{"role bukan string", sign(t, map[string]any{"role": 5}), true, AIProviders{}, img, 401, `{"error":"Unauthorized: Invalid or expired token"}`},
		{"praktikan", sign(t, map[string]any{"role": "praktikan"}), true, AIProviders{}, img, 403, `{"error":"Forbidden: Hanya staf/asisten yang dapat melakukan analisis jadwal KRS"}`},
		{"alias admin ditolak", sign(t, map[string]any{"role": "admin"}), true, AIProviders{}, img, 403, `{"error":"Forbidden: Hanya staf/asisten yang dapat melakukan analisis jadwal KRS"}`},
		{"tanpa file", bearer, true, AIProviders{}, dto.AnalyzeScheduleRequest{MimeType: "image/png"}, 400, `{"error":"fileBase64 and mimeType are required"}`},
		{"key gemini kosong", bearer, true, AIProviders{Bynara: &fakeBynara{}}, img, 500, `{"error":"GEMINI_API_KEY belum dikonfigurasi di server"}`},
		{"key bynara kosong", bearer, true, AIProviders{Gemini: &fakeGemini{}}, pdfStepfun, 500, `{"error":"BYNARA_API_KEY belum dikonfigurasi di Vercel Environment Variables"}`},
		{"stepfun pdf", bearer, true, AIProviders{Bynara: &fakeBynara{}}, pdfStepfun, 400, `{"error":"Model Stepfun 3.7 Flash hanya mendukung file gambar (JPG/PNG). Untuk file PDF silakan gunakan model Gemini."}`},
		{"gemini 502", bearer, true, AIProviders{Gemini: &fakeGemini{err: &gemini.APIError{Status: 429, Body: "quota"}}}, img, 502, `{"error":"Gemini API error (429)"}`},
		{"bynara 502", bearer, true, AIProviders{Bynara: &fakeBynara{err: &bynara.APIError{Status: 503, Body: long}}}, stepfun, 502, `{"error":"Bynara API error (503): ` + strings.Repeat("é", 150) + `"}`},
		{"jaringan putus", bearer, true, AIProviders{Gemini: &fakeGemini{err: errors.New("dial tcp: timeout")}}, img, 500, `{"error":"dial tcp: timeout"}`},
		{"format tidak bisa diproses", bearer, true, AIProviders{Gemini: &fakeGemini{text: long}}, img, 422, `{"error":"` + slotsUnparsableText + `","raw":"` + strings.Repeat("é", 300) + `"}`},
		{"respons kosong", bearer, true, AIProviders{Gemini: &fakeGemini{}}, img, 422, `{"error":"` + slotsUnparsableText + `","raw":""}`},
		{"bukan array", bearer, true, AIProviders{Gemini: &fakeGemini{text: `{"a":1}`}}, img, 422, `{"error":"` + slotsUnparsableText + `","raw":"{\"a\":1}"}`},
	}
	for _, tc := range cases {
		uc := NewScheduleAIUsecase(NewTokens(testSecret), tc.ai)
		_, f := uc.Analyze(context.Background(), tc.token, tc.hasBearer, tc.req)
		if f == nil {
			t.Errorf("%s: tidak gagal", tc.name)
			continue
		}
		if got := string(mustJSON(t, f.Body)); f.Status != tc.status || got != tc.body {
			t.Errorf("%s: %d %s", tc.name, f.Status, got)
		}
	}
}

func TestStepfunPDFSendsNoRequest(t *testing.T) {
	b := &fakeBynara{}
	uc := NewScheduleAIUsecase(NewTokens(testSecret), AIProviders{Bynara: b})
	_, f := uc.Analyze(context.Background(), staffBearer(t), true,
		dto.AnalyzeScheduleRequest{FileBase64: "x", MimeType: "application/pdf", Model: "stepfun-3.7-flash"})
	if f == nil || f.Status != http.StatusBadRequest || len(b.calls) != 0 {
		t.Fatalf("failure %+v, panggilan %d", f, len(b.calls))
	}
}

func TestExtractSlots(t *testing.T) {
	cases := []struct {
		text, want string
	}{
		{slotsJSON, slotsJSON},
		{"```json\n[]\n```", "[]"},
		{"[1, 2]", "[1,2]"},
		{"Jawaban:\n[ {\"a\": 1},\n {\"b\": 2} ]\nSelesai", `[{"a":1},{"b":2}]`},
		{"null", ""},
		{"", ""},
		{"[{rusak}]", ""},
	}
	for _, tc := range cases {
		got := ""
		if slots := extractSlots(tc.text); slots != nil {
			got = string(mustJSON(t, slots))
		}
		if got != tc.want {
			t.Errorf("%q: %s, ingin %s", tc.text, got, tc.want)
		}
	}
}

func TestJSSubstring(t *testing.T) {
	if got := jsSubstring("a😀b", 3); got != "a😀" {
		t.Fatalf("surrogate pair dihitung dua unit: %q", got)
	}
	if got := jsSubstring("abc", 10); got != "abc" {
		t.Fatalf("%q", got)
	}
}
