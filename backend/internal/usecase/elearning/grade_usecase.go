package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
)

const (
	defaultGradeModel = "gpt-os-120b"
	gradeConcurrency  = 3
)

// Pesan per attempt dikirim apa adanya di results[].error, jadi teksnya mengikuti api/grade.ts (huruf besar, titik).
type gradeError string

func (e gradeError) Error() string { return string(e) }

type GradeUsecase struct {
	auth        *AuthUsecase
	users       repository.UserRepository
	assessments repository.AssessmentRepository
	ai          AIClients
	now         func() time.Time
}

func NewGradeUsecase(auth *AuthUsecase, users repository.UserRepository, assessments repository.AssessmentRepository, ai AIClients, now func() time.Time) *GradeUsecase {
	return &GradeUsecase{auth: auth, users: users, assessments: assessments, ai: ai, now: now}
}

func (uc *GradeUsecase) Grade(ctx context.Context, token string, req *dto.GradeRequest) (*dto.GradeResponse, error) {
	if token == "" {
		return nil, Unauthorized("Unauthorized: Missing token")
	}
	claims, err := uc.auth.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}
	graderNIM := pick(claims, nil, "nim", "sub")
	email := claims["email"]

	// Error DB ditelan seperti roleError di api/grade.ts: hasilnya sama dengan profil tidak ditemukan.
	var role string
	found := false
	if truthy(graderNIM) {
		role, err = uc.users.RoleByNIM(ctx, jsString(graderNIM))
		found = err == nil
	}
	if !found && truthy(email) {
		role, err = uc.users.RoleByEmail(ctx, jsString(email))
		found = err == nil
	}
	if !found {
		return nil, Forbidden("Akses Ditolak: Profil Anda tidak ditemukan (NIM: " + jsToString(or(graderNIM, "N/A")) + ", Email: " + jsToString(or(email, "N/A")) + ").")
	}
	// Dibandingkan dengan users.role mentah seperti kode lama, jadi nilai kanonik "koordinator" ikut ditolak.
	if role != RoleAdmin && role != RoleKordas && role != RoleAsisten {
		return nil, Forbidden("Akses Ditolak: Hanya Asisten, Kordas, atau Admin yang bisa menilai.")
	}

	var ids []any
	if len(req.AttemptIDs) == 0 || json.Unmarshal(req.AttemptIDs, &ids) != nil || len(ids) == 0 {
		return nil, BadRequest("Invalid payload: attemptIds array is required")
	}
	model := requestedModel(req.RequestedModel)

	results := make([]dto.GradeResult, len(ids))
	sem := make(chan struct{}, gradeConcurrency)
	var wg sync.WaitGroup
	for i, id := range ids {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			results[i] = uc.gradeAttempt(ctx, id, model)
		})
	}
	wg.Wait()

	graded := 0
	for i := range results {
		if results[i].Success {
			graded++
		}
	}
	return &dto.GradeResponse{Success: true, GradedCount: graded, Results: results}, nil
}

// Default destructuring JS hanya berlaku saat field tidak dikirim; null atau tipe lain tidak cocok dengan model mana pun.
func requestedModel(raw json.RawMessage) string {
	if len(raw) == 0 {
		return defaultGradeModel
	}
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

func (uc *GradeUsecase) gradeAttempt(ctx context.Context, id any, model string) dto.GradeResult {
	res, err := uc.gradeAttemptErr(ctx, id, model)
	if err != nil {
		log.Printf("elearning grade: attempt %v gagal: %v", id, err)
		msg := err.Error()
		if msg == "" {
			msg = "Error internal penilaian AI"
		}
		return dto.GradeResult{AttemptID: id, Success: false, Error: msg}
	}
	return *res
}

func (uc *GradeUsecase) gradeAttemptErr(ctx context.Context, id any, model string) (*dto.GradeResult, error) {
	attemptID := jsToString(id)
	attempt, err := uc.assessments.Attempt(ctx, attemptID)
	if err != nil {
		return nil, gradeError("Attempt tidak ditemukan: " + dbMessage(err))
	}
	if attempt.Status != "submitted" && attempt.Status != "graded" {
		return nil, gradeError("Attempt status bukan 'submitted' atau 'graded' (status saat ini: " + attempt.Status + ")")
	}

	studentName, _ := uc.users.NamaByNIM(ctx, attempt.NIM)
	if studentName == "" {
		studentName = "Mahasiswa"
	}
	var rules any
	if raw, err := uc.assessments.GradingRules(ctx, attempt.MenuType); err == nil && len(raw) > 0 {
		rules, _ = parseJS(raw)
	}
	if !jsTruthy(rules) {
		rules = nil
	}

	questionIDs := jsArray(attempt.SelectedQuestions)
	if len(questionIDs) == 0 {
		return nil, gradeError("Attempt tidak memiliki soal terasosiasi.")
	}
	keys := make([]string, len(questionIDs))
	for i, q := range questionIDs {
		keys[i] = jsToString(q)
	}
	questions, err := uc.assessments.Questions(ctx, keys)
	if err != nil {
		return nil, gradeError("Gagal memuat soal: " + dbMessage(err))
	}
	if len(questions) == 0 {
		return nil, gradeError("Gagal memuat soal: undefined")
	}

	prompt, err := buildGradePrompt(attempt, studentName, rules, questionIDs, questions)
	if err != nil {
		return nil, err
	}
	text, err := uc.callAI(ctx, model, prompt)
	if err != nil {
		return nil, err
	}
	aiGrades, finalScore, err := parseGradeResult(text)
	if err != nil {
		return nil, err
	}
	if err := uc.assessments.SaveGrade(ctx, attemptID, aiGrades, finalScore, uc.now().UTC()); err != nil {
		return nil, gradeError(dbMessage(err))
	}
	return &dto.GradeResult{AttemptID: id, Success: true, FinalScore: finalScore, StudentName: studentName}, nil
}

func jsArray(raw []byte) []any {
	if len(raw) == 0 {
		return nil
	}
	v, err := parseJS(raw)
	if err != nil {
		return nil
	}
	arr, _ := v.([]any)
	return arr
}

func (uc *GradeUsecase) callAI(ctx context.Context, model, prompt string) (string, error) {
	if model == defaultGradeModel && uc.ai.Ollama != nil {
		return uc.ai.Ollama.NativeChatJSON(ctx, prompt)
	}
	if len(uc.ai.GeminiKeys) == 0 {
		return "", gradeError("API Keys Gemini tidak dikonfigurasi di server.")
	}
	geminiModel := "gemini-2.0-flash-exp"
	if model == "gemini-2.5-flash" {
		geminiModel = model
	}
	parts := []gemini.Part{{Text: prompt}}
	cfg := gemini.GenerationConfig{ResponseMimeType: "application/json"}
	var text string
	var lastErr error
	for _, key := range uc.ai.rotation() {
		t, err := uc.ai.Gemini.GenerateParts(ctx, key, geminiModel, parts, cfg)
		if err == nil {
			text = t
			break
		}
		lastErr = err
	}
	// lastErr sengaja tidak di-reset saat sukses: teks kosong setelah percobaan gagal tetap melempar error lama.
	if text == "" && lastErr != nil {
		return "", lastErr
	}
	return text, nil
}

// Setara JSON.parse lalu `grades || {}` dan `total_overall_score || 0` di api/grade.ts.
func parseGradeResult(text string) (json.RawMessage, any, error) {
	clean := strings.TrimSpace(text)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimSpace(strings.TrimSuffix(clean, "```"))
	dec := json.NewDecoder(strings.NewReader(clean))
	dec.UseNumber()
	var top any
	if err := dec.Decode(&top); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, gradeError("Unexpected end of JSON input")
		}
		return nil, nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, nil, errJSONTrailing
	}
	if top == nil {
		return nil, nil, gradeError("Cannot read properties of null (reading 'grades')")
	}
	obj, _ := top.(map[string]any)
	grades := any(map[string]any{})
	if jsTruthy(obj["grades"]) {
		grades = obj["grades"]
	}
	raw, err := json.Marshal(grades)
	if err != nil {
		return nil, nil, err
	}
	var score any = json.Number("0")
	if jsTruthy(obj["total_overall_score"]) {
		score = obj["total_overall_score"]
	}
	return raw, score, nil
}

const fence = "```"

const gradePromptFooter = `TUGAS ANDA:
1. Evaluasi masing-masing jawaban soal di atas berdasarkan kriteria rubrik yang diberikan.
2. Berikan analisis logis, poin positif, saran perbaikan, dan feedback ringkas dalam bahasa Indonesia.
3. Kembalikan respons Anda hanya dalam bentuk format JSON terstruktur yang valid, tanpa teks penjelasan tambahan di luar JSON.

Format respons JSON yang harus Anda hasilkan:
{
  "grades": {
    "[soalId_1]": {
      "scores": {
        "correctness_or_compilation": 0, // Nilai berdasarkan rubrik
        "explanation_or_instruction": 0,
        "completion_or_timing": 0
      },
      "total_score": 0, // Jumlah dari sub-scores
      "feedback": "Komentar spesifik soal ini..."
    },
    "[soalId_2]": {
      ...
    }
  },
  "total_overall_score": 0, // Total nilai dari seluruh soal asesmen
  "overall_feedback": "Komentar umum pengerjaan asesmen mahasiswa ini..."
}`

// Teks prompt disalin verbatim dari api/grade.ts langkah E supaya hasil penilaian AI tidak bergeser.
func buildGradePrompt(attempt *entity.AssessmentAttempt, studentName string, rules any, questionIDs []any, questions []entity.AssessmentQuestion) (string, error) {
	var b strings.Builder
	b.WriteString(`Anda adalah Asisten Praktikum AI. Anda ditugaskan menilai pengerjaan ujian mahasiswa bernama "` + studentName + `" (NIM: ` + attempt.NIM + `).
Tipe Asesmen: ` + strings.ToUpper(attempt.MenuType) + `
Evaluasi seluruh soal berikut dan berikan skor sesuai kriteria penilaian.

Berikut adalah daftar soal dan jawaban draf mahasiswa:
======================================================
`)
	var answers any
	if v, err := parseJS(attempt.Answers); err == nil {
		answers = v
	}
	for idx := range questions {
		q := &questions[idx]
		n := strconv.Itoa(idx + 1)
		answer := or(prop(answers, q.ID), &jsObject{vals: map[string]any{}})
		b.WriteString(`SOAL ` + n + ` (ID Soal: ` + q.ID + `):
- Judul Soal: ` + jsToString(nullable(q.Title)) + `
- Instruksi Soal: ` + jsToString(nullable(q.Instruction)) + `
- Tipe Soal: ` + jsToString(nullable(q.Type)) + `
- Kriteria Kesulitan: ` + jsToString(nullable(q.Difficulty)) + `
- Modul Terkait: Modul ` + jsToString(or(nullable(q.ModuleAssociation), "N/A")) + `
- Kode Awal / Template: 
` + fence + `
` + jsToString(or(nullable(q.InitialCode), "")) + `
` + fence + `
- Kunci Jawaban Referensi / Solusi:
` + fence + `
` + jsToString(or(nullable(q.ReferenceSolution), "")) + `
` + fence + `

JAWABAN MAHASISWA (SOAL ` + n + `):
- Teks Jawaban / Penjelasan: "` + jsToString(or(prop(answer, "answerText"), "KOSONG")) + `"
- Kode yang Dikirim:
` + fence + `
` + jsToString(or(prop(answer, "codeSubmitted"), "KOSONG")) + `
` + fence + `
- Hasil Output Sandbox (Stdout): "` + jsToString(or(prop(answer, "outputStandard"), "")) + `"
- Error Sandbox (Stderr): "` + jsToString(or(prop(answer, "errors"), "")) + `"

KRITERIA SCORING KHUSUS UNTUK SOAL INI:
`)
		rubric, err := rubricLine(attempt.MenuType, rules, q, questionIDs)
		if err != nil {
			return "", err
		}
		b.WriteString(rubric)
		b.WriteString("======================================================\n\n")
	}
	b.WriteString(gradePromptFooter)
	return b.String(), nil
}

func rubricLine(menuType string, rules any, q *entity.AssessmentQuestion, questionIDs []any) (string, error) {
	difficulty := nullable(q.Difficulty)
	diffKey := jsToString(difficulty)
	qType := jsToString(nullable(q.Type))
	hasRules := rules != nil

	switch {
	case hasRules && (menuType == "pre_test" || menuType == "post_test") && jsTruthy(prop(prop(rules, "difficulties"), diffKey)):
		if difficulty == nil {
			return "", gradeError("Cannot read properties of null (reading 'toUpperCase')")
		}
		diffRules := prop(prop(rules, "difficulties"), diffKey)
		return "- " + strings.ToUpper(diffKey) + ": " + criteriaString(prop(diffRules, "criteria")) + "\n", nil
	case hasRules && menuType == "program_keterampilan" && jsTruthy(prop(rules, "criteria")):
		list, ok := prop(rules, "criteria").([]any)
		if !ok {
			return "", gradeError("dynamicRules.criteria.map is not a function")
		}
		items := make([]string, len(list))
		for i, c := range list {
			if c == nil {
				return "", gradeError("Cannot read properties of null (reading 'label')")
			}
			items[i] = jsToString(prop(c, "label")) + ": " + jsToString(prop(c, "nilai")) + " poin"
		}
		return "- Program Keterampilan: " + strings.Join(items, ", ") + " (Total " + jsToString(or(prop(rules, "total_max_score"), 85.0)) + ").\n", nil
	case hasRules && menuType == "ujian_praktik":
		soalKey := "soal_" + strconv.Itoa(indexOfString(questionIDs, q.ID)+1)
		if specific := prop(rules, soalKey); jsTruthy(specific) {
			label := strings.Replace(strings.ToUpper(soalKey), "_", " ", 1)
			return "- Ujian Praktik " + label + ": " + criteriaString(prop(specific, "criteria")) + " (Maks " + jsToString(or(prop(specific, "max_score"), 0.0)) + " poin).\n", nil
		}
		if flow := prop(rules, "soal_6_flowchart"); qType == "flowchart_translation" && jsTruthy(flow) {
			return "- Flowchart to Program: " + criteriaString(prop(flow, "criteria")) + ".\n", nil
		}
		if s15 := prop(rules, "soal_1_5"); jsTruthy(s15) {
			return "- Ujian Praktik Soal 1-5: " + criteriaString(prop(s15, "criteria")) + ".\n", nil
		}
		return "", nil
	}
	return staticRubric(menuType, diffKey, qType), nil
}

func staticRubric(menuType, difficulty, qType string) string {
	switch menuType {
	case "pre_test":
		switch difficulty {
		case "easy":
			return "- Pilihan Ganda / Jawaban Singkat Easy: Jawaban Benar = 20 poin, Jawaban Salah = 8 poin, Kosong = 0 poin.\n"
		case "medium":
			return "- Jawaban Singkat + Penjelasan Medium: Jawaban Benar Singkat = 10 poin, Jawaban Benar + Penjelasan Tepat = 15 poin, Jawaban Salah + Penjelasan Logis = 7 poin, Jawaban Salah = 3 poin, Kosong = 0 poin.\n"
		case "hard":
			return "- Jawaban Singkat + Penjelasan Hard: Jawaban Benar Singkat = 15 poin, Jawaban Benar + Penjelasan Tepat = 25 poin, Jawaban Salah + Penjelasan Logis = 10 poin, Jawaban Salah = 5 poin, Kosong = 0 poin.\n"
		}
	case "post_test":
		switch difficulty {
		case "easy":
			return "- Easy: Jawaban Benar = 20 poin, Jawaban Salah = 8 poin, Kosong = 0 poin.\n"
		case "medium":
			return "- Medium: Jawaban Benar Singkat = 20 poin, Jawaban Benar + Penjelasan = 35 poin, Jawaban Salah + Penjelasan Logis = 15 poin, Jawaban Salah = 10 poin, Kosong = 0 poin.\n"
		case "hard":
			return "- Hard Coding: Dapat Berjalan Tanpa Error = 7 poin, Sesuai Petunjuk = 25 poin, Selesai/Kriteria output terpenuhi = 13 poin, Belum Selesai = 3 poin.\n"
		}
	case "program_keterampilan":
		return "- Program Keterampilan: Dapat Berjalan Tanpa Error = 30 poin (jika berjalan sebagian = 15), Sesuai Petunjuk/Instruksi Sintaks = 35 poin, Selesai/Tepat Waktu = 20 poin, Belum Selesai = 10 poin (Total 85).\n"
	case "ujian_praktik":
		if qType == "flowchart_translation" {
			return "- Flowchart to Program: Dapat Berjalan Tanpa Error = 7 poin, Kesesuaian Alur & Logika Flowchart = 25 poin, Selesai/Tepat Waktu = 13 poin, Belum Selesai = 3 poin.\n"
		}
		return "- Ujian Praktik Soal 1-5 (Hard Coding): Dapat Berjalan Tanpa Error = 7 poin, Sesuai Petunjuk = 25 poin, Selesai/Tepat Waktu = 13 poin, Belum Selesai = 3 poin.\n"
	}
	return ""
}

// Object.entries(x || {}).map(([k, v]) => `${k.replace(/_/g, ' ')}: ${v} poin`).join(', ')
func criteriaString(criteria any) string {
	var items []string
	add := func(k string, v any) {
		items = append(items, strings.ReplaceAll(k, "_", " ")+": "+jsToString(v)+" poin")
	}
	switch c := criteria.(type) {
	case *jsObject:
		for _, k := range c.entries() {
			add(k, c.vals[k])
		}
	case []any:
		for i, v := range c {
			add(strconv.Itoa(i), v)
		}
	case string:
		for i, r := range []rune(c) {
			add(strconv.Itoa(i), string(r))
		}
	}
	return strings.Join(items, ", ")
}

// Array.prototype.indexOf memakai ===, jadi hanya elemen string yang bisa cocok dengan id soal.
func indexOfString(list []any, s string) int {
	for i, v := range list {
		if str, ok := v.(string); ok && str == s {
			return i
		}
	}
	return -1
}
