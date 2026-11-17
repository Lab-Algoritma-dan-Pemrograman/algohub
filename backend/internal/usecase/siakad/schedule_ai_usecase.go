package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf16"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/bynara"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
)

const (
	stepfunModel        = "stepfun-3.7-flash"
	geminiDefaultModel  = "gemini-3-flash"
	geminiAltModel      = "gemini-2.5-flash"
	slotsUnparsableText = "AI berhasil merespons tetapi format data tidak dapat diproses. Pastikan gambar KRS/Jadwal terlihat jelas."
)

// Disalin byte-per-byte dari template literal kode lama (termasuk spasi di akhir baris)
// karena hasil model bergantung pada teks prompt.
//
//nolint:misspell // teks berbahasa Indonesia, bukan salah eja
const schedulePrompt = "\n" +
	"    Anda adalah sistem penjadwalan cerdas. \n" +
	"    Berikut adalah gambar/dokumen KRS atau Jadwal Kuliah seorang mahasiswa.\n" +
	"    Tugas Anda:\n" +
	"    1. Jam operasional asisten laboratorium adalah hari Senin sampai Sabtu, dari pukul 06:00 hingga 22:00.\n" +
	"    2. Analisis jadwal tersebut dan carilah JAM KOSONG (waktu di mana mahasiswa tersebut TIDAK ADA jadwal kuliah) dalam rentang jam operasional tersebut (06:00 - 22:00).\n" +
	"    3. Abaikan hari Minggu.\n" +
	"    4. Jika ada rentang waktu kosong yang saling berurutan di hari yang sama, gabungkan rentang tersebut menjadi satu waktu. (Contoh: kosong jam 06:00-10:00 dan 10:00-12:00 digabung jadi 06:00-12:00).\n" +
	"    \n" +
	"    Format jam harus hh:mm (contoh: 06:00, 13:30, 22:00).\n" +
	"\n" +
	"    KEMBALIKAN HANYA ARRAY JSON VALID TANPA TEKS LAIN ATAU MARKDOWN BACKTICKS!\n" +
	"    Contoh format output wajib:\n" +
	"    [\n" +
	"        {\"day_of_week\": \"Senin\", \"start_time\": \"06:00\", \"end_time\": \"12:00\"},\n" +
	"        {\"day_of_week\": \"Rabu\", \"start_time\": \"13:00\", \"end_time\": \"22:00\"}\n" +
	"    ]\n" +
	"  "

type GeminiClient interface {
	GenerateContent(ctx context.Context, model, prompt, mimeType, dataBase64 string) (string, error)
}

type BynaraClient interface {
	ChatImage(ctx context.Context, model, prompt, dataURL string) (string, error)
}

// Klien bernilai nil berarti API key-nya tidak dikonfigurasi; kode lama membalas 500 per request
// alih-alih gagal start.
type AIProviders struct {
	Gemini GeminiClient
	Bynara BynaraClient
}

type ScheduleAIUsecase struct {
	tokens *Tokens
	ai     AIProviders
}

func NewScheduleAIUsecase(tokens *Tokens, ai AIProviders) *ScheduleAIUsecase {
	return &ScheduleAIUsecase{tokens: tokens, ai: ai}
}

// T-15: hanya header Bearer dan role mentah tanpa alias, sama dengan kode lama.
func (u *ScheduleAIUsecase) Analyze(ctx context.Context, token string, hasBearer bool, req dto.AnalyzeScheduleRequest) (map[string]any, *Failure) {
	if !hasBearer {
		return nil, fail(http.StatusUnauthorized, "Unauthorized: Missing session token")
	}
	claims, err := u.tokens.Verify(token)
	if err != nil {
		return nil, fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	role := ""
	if r := claims["role"]; truthy(r) {
		s, ok := r.(string)
		if !ok {
			// Kode lama memanggil toLowerCase di dalam blok verifikasi, jadi role bukan string = token invalid.
			return nil, fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
		}
		role = strings.ToLower(s)
	}
	if role != "asisten" && role != "koordinator" {
		return nil, fail(http.StatusForbidden, "Forbidden: Hanya staf/asisten yang dapat melakukan analisis jadwal KRS")
	}
	if req.FileBase64 == "" || req.MimeType == "" {
		return nil, fail(http.StatusBadRequest, "fileBase64 and mimeType are required")
	}

	text, f := u.generate(ctx, req)
	if f != nil {
		return nil, f
	}
	slots := extractSlots(text)
	if slots == nil {
		return nil, &Failure{Status: http.StatusUnprocessableEntity, Body: map[string]any{
			"error": slotsUnparsableText,
			"raw":   jsSubstring(text, 300),
		}}
	}
	return map[string]any{"slots": slots}, nil
}

func (u *ScheduleAIUsecase) generate(ctx context.Context, req dto.AnalyzeScheduleRequest) (string, *Failure) {
	if req.Model == stepfunModel {
		if u.ai.Bynara == nil {
			return "", fail(http.StatusInternalServerError, "BYNARA_API_KEY belum dikonfigurasi di Vercel Environment Variables")
		}
		if req.MimeType == "application/pdf" {
			return "", fail(http.StatusBadRequest, "Model Stepfun 3.7 Flash hanya mendukung file gambar (JPG/PNG). Untuk file PDF silakan gunakan model Gemini.")
		}
		text, err := u.ai.Bynara.ChatImage(ctx, stepfunModel, schedulePrompt, "data:"+req.MimeType+";base64,"+req.FileBase64)
		var apiErr *bynara.APIError
		if errors.As(err, &apiErr) {
			return "", fail(http.StatusBadGateway, fmt.Sprintf("Bynara API error (%d): %s", apiErr.Status, jsSubstring(apiErr.Body, 150)))
		}
		return text, aiFailure(err)
	}

	if u.ai.Gemini == nil {
		return "", fail(http.StatusInternalServerError, "GEMINI_API_KEY belum dikonfigurasi di server")
	}
	model := geminiDefaultModel
	if req.Model == geminiAltModel {
		model = geminiAltModel
	}
	text, err := u.ai.Gemini.GenerateContent(ctx, model, schedulePrompt, req.MimeType, req.FileBase64)
	var apiErr *gemini.APIError
	if errors.As(err, &apiErr) {
		return "", fail(http.StatusBadGateway, fmt.Sprintf("Gemini API error (%d)", apiErr.Status))
	}
	return text, aiFailure(err)
}

func aiFailure(err error) *Failure {
	if err == nil {
		return nil
	}
	return fail(http.StatusInternalServerError, err.Error())
}

var (
	jsonFenceRe = regexp.MustCompile("(?i)```json")
	slotArrayRe = regexp.MustCompile(`(?s)\[\s*\{.*\}\s*\]`)
)

// Model sering membungkus JSON dengan markdown atau teks penjelas walau dilarang prompt,
// jadi dicoba dulu setelah fence dibuang, lalu potongan array objek pertama sampai terakhir.
func extractSlots(text string) []json.RawMessage {
	if text == "" {
		return nil
	}
	clean := strings.TrimSpace(strings.ReplaceAll(jsonFenceRe.ReplaceAllString(text, ""), "```", ""))
	var slots []json.RawMessage
	if json.Unmarshal([]byte(clean), &slots) == nil && slots != nil {
		return slots
	}
	if m := slotArrayRe.FindString(text); m != "" {
		if json.Unmarshal([]byte(m), &slots) == nil && slots != nil {
			return slots
		}
	}
	return nil
}

// String.prototype.substring menghitung unit UTF-16, bukan byte atau rune.
func jsSubstring(s string, n int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= n {
		return s
	}
	return string(utf16.Decode(units[:n]))
}
