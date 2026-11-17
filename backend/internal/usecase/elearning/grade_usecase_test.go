package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

var gradeNow = time.Date(2026, 11, 17, 8, 30, 0, 0, time.UTC)

type fakeOllama struct {
	mu      sync.Mutex
	prompts []string
	text    string
	err     error
}

func (f *fakeOllama) NativeChatJSON(_ context.Context, prompt string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prompts = append(f.prompts, prompt)
	return f.text, f.err
}

type gradeFixture struct {
	uc          *usecase.GradeUsecase
	users       *mocks.UserRepository
	assessments *mocks.AssessmentRepository
}

func newGrade(t *testing.T, ai usecase.AIClients) *gradeFixture {
	t.Helper()
	f := &gradeFixture{users: mocks.NewUserRepository(t), assessments: mocks.NewAssessmentRepository(t)}
	if ai.Shuffle == nil {
		ai.Shuffle = reverseKeys
	}
	f.uc = usecase.NewGradeUsecase(newAuth(testConfig(), nil), f.users, f.assessments, ai, func() time.Time { return gradeNow })
	return f
}

func graderToken(t *testing.T) string {
	return supabaseToken(t, jwt.MapClaims{"nim": "199001", "email": "asisten@lab.test"})
}

func gradeReq(ids string) *dto.GradeRequest {
	return &dto.GradeRequest{AttemptIDs: json.RawMessage(ids)}
}

func strp(s string) *string { return &s }

func submitted(id, menu string) *entity.AssessmentAttempt {
	return &entity.AssessmentAttempt{
		ID: id, NIM: "202711001", MenuType: menu, Status: "submitted",
		SelectedQuestions: []byte(`["q1"]`),
		Answers:           []byte(`{"q1":{"answerText":"jawab","codeSubmitted":"print(1)","outputStandard":"1"}}`),
	}
}

func easyQuestion() entity.AssessmentQuestion {
	return entity.AssessmentQuestion{ID: "q1", Title: strp("Soal A"), Instruction: strp("Kerjakan"), Type: strp("essay"), Difficulty: strp("easy")}
}

func (f *gradeFixture) expectAttempt(a *entity.AssessmentAttempt, rules string, qs ...entity.AssessmentQuestion) {
	f.assessments.On("Attempt", mock.Anything, a.ID).Return(a, nil)
	f.users.On("NamaByNIM", mock.Anything, a.NIM).Return("Budi", nil)
	if rules == "" {
		f.assessments.On("GradingRules", mock.Anything, a.MenuType).Return(nil, gorm.ErrRecordNotFound)
	} else {
		f.assessments.On("GradingRules", mock.Anything, a.MenuType).Return(json.RawMessage(rules), nil)
	}
	f.assessments.On("Questions", mock.Anything, []string{"q1"}).Return(qs, nil)
}

func TestGradeAuthAndProfile(t *testing.T) {
	ctx := context.Background()
	t.Run("tanpa token", func(t *testing.T) {
		_, err := newGrade(t, usecase.AIClients{}).uc.Grade(ctx, "", gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Missing token")
	})
	t.Run("token rusak", func(t *testing.T) {
		_, err := newGrade(t, usecase.AIClients{}).uc.Grade(ctx, "rusak", gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")
	})
	t.Run("profil tidak ditemukan", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		f.users.On("RoleByNIM", mock.Anything, "199001").Return("", gorm.ErrRecordNotFound)
		f.users.On("RoleByEmail", mock.Anything, "asisten@lab.test").Return("", errors.New("db mati"))
		_, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusForbidden, "Akses Ditolak: Profil Anda tidak ditemukan (NIM: 199001, Email: asisten@lab.test).")
	})
	t.Run("profil tidak ditemukan tanpa nim dan email", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		_, err := f.uc.Grade(ctx, supabaseToken(t, jwt.MapClaims{"role": "authenticated"}), gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusForbidden, "Akses Ditolak: Profil Anda tidak ditemukan (NIM: N/A, Email: N/A).")
	})
	t.Run("nim dari sub", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		f.users.On("RoleByNIM", mock.Anything, "199002").Return("mahasiswa", nil)
		_, err := f.uc.Grade(ctx, supabaseToken(t, jwt.MapClaims{"sub": "199002"}), gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusForbidden, "Akses Ditolak: Hanya Asisten, Kordas, atau Admin yang bisa menilai.")
	})
	t.Run("role mentah koordinator ditolak", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		f.users.On("RoleByNIM", mock.Anything, "199001").Return("koordinator", nil)
		_, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a"]`))
		assertUCError(t, err, http.StatusForbidden, "Akses Ditolak: Hanya Asisten, Kordas, atau Admin yang bisa menilai.")
	})
	for _, ids := range []string{"", `[]`, `"a"`, `null`} {
		t.Run("attemptIds "+ids, func(t *testing.T) {
			f := newGrade(t, usecase.AIClients{})
			f.users.On("RoleByNIM", mock.Anything, "199001").Return("", gorm.ErrRecordNotFound)
			f.users.On("RoleByEmail", mock.Anything, "asisten@lab.test").Return("asisten", nil)
			_, err := f.uc.Grade(ctx, graderToken(t), gradeReq(ids))
			assertUCError(t, err, http.StatusBadRequest, "Invalid payload: attemptIds array is required")
		})
	}
}

func asisten(f *gradeFixture) {
	f.users.On("RoleByNIM", mock.Anything, "199001").Return("asisten", nil)
}

func TestGradeOllamaBranchAndSave(t *testing.T) {
	oll := &fakeOllama{text: "```json\n{\"grades\":{\"q1\":{\"total_score\":20}},\"total_overall_score\":85}\n```"}
	g := &fakeGemini{}
	f := newGrade(t, usecase.AIClients{Gemini: g, GeminiKeys: []string{"k1"}, Ollama: oll})
	asisten(f)
	f.expectAttempt(submitted("a1", "pre_test"), "", easyQuestion())
	f.assessments.On("SaveGrade", mock.Anything, "a1", mock.MatchedBy(func(raw json.RawMessage) bool {
		return string(raw) == `{"q1":{"total_score":20}}`
	}), json.Number("85"), gradeNow).Return(nil)

	res, err := f.uc.Grade(context.Background(), graderToken(t), gradeReq(`["a1"]`))
	require.NoError(t, err)
	assert.Equal(t, &dto.GradeResponse{Success: true, GradedCount: 1, Results: []dto.GradeResult{
		{AttemptID: "a1", Success: true, FinalScore: json.Number("85"), StudentName: "Budi"},
	}}, res)
	assert.Len(t, oll.prompts, 1)
	assert.Empty(t, g.calls)
}

func TestGradeGeminiBranch(t *testing.T) {
	cases := []struct {
		name, model, wantModel string
		ollama                 bool
	}{
		{"default tanpa ollama", "", "gemini-2.0-flash-exp", false},
		{"gemini-2.5-flash", `"gemini-2.5-flash"`, "gemini-2.5-flash", true},
		{"null bukan default", `null`, "gemini-2.0-flash-exp", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &fakeGemini{text: `{"grades":null}`}
			ai := usecase.AIClients{Gemini: g, GeminiKeys: []string{"k1"}}
			oll := &fakeOllama{}
			if tc.ollama {
				ai.Ollama = oll
			}
			f := newGrade(t, ai)
			asisten(f)
			f.expectAttempt(submitted("a1", "pre_test"), "", easyQuestion())
			f.assessments.On("SaveGrade", mock.Anything, "a1", json.RawMessage(`{}`), json.Number("0"), gradeNow).Return(nil)
			req := gradeReq(`["a1"]`)
			req.RequestedModel = json.RawMessage(tc.model)

			res, err := f.uc.Grade(context.Background(), graderToken(t), req)
			require.NoError(t, err)
			assert.True(t, res.Results[0].Success, res.Results[0].Error)
			require.Len(t, g.calls, 1)
			assert.Equal(t, tc.wantModel, g.calls[0].model)
			assert.Equal(t, "application/json", g.calls[0].cfg.ResponseMimeType)
			assert.Empty(t, oll.prompts)
		})
	}
}

func TestGradeAttemptErrors(t *testing.T) {
	ctx := context.Background()
	t.Run("status salah", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		asisten(f)
		a := submitted("a1", "pre_test")
		a.Status = "in_progress"
		f.assessments.On("Attempt", mock.Anything, "a1").Return(a, nil)
		res, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a1"]`))
		require.NoError(t, err)
		assert.Equal(t, dto.GradeResult{AttemptID: "a1", Error: "Attempt status bukan 'submitted' atau 'graded' (status saat ini: in_progress)"}, res.Results[0])
		assert.Equal(t, 0, res.GradedCount)
	})
	t.Run("attempt tidak ada", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		asisten(f)
		f.assessments.On("Attempt", mock.Anything, "a9").Return(nil, gorm.ErrRecordNotFound)
		res, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a9"]`))
		require.NoError(t, err)
		assert.Equal(t, "Attempt tidak ditemukan: Cannot coerce the result to a single JSON object", res.Results[0].Error)
	})
	t.Run("tanpa kunci gemini", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		asisten(f)
		f.expectAttempt(submitted("a1", "pre_test"), "", easyQuestion())
		res, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a1"]`))
		require.NoError(t, err)
		assert.Equal(t, "API Keys Gemini tidak dikonfigurasi di server.", res.Results[0].Error)
	})
	t.Run("soal kosong", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{})
		asisten(f)
		f.expectAttempt(submitted("a1", "pre_test"), "")
		res, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a1"]`))
		require.NoError(t, err)
		assert.Equal(t, "Gagal memuat soal: undefined", res.Results[0].Error)
	})
	t.Run("ollama error", func(t *testing.T) {
		f := newGrade(t, usecase.AIClients{Ollama: &fakeOllama{err: errors.New("Ollama API returned status: 502")}})
		asisten(f)
		f.expectAttempt(submitted("a1", "pre_test"), "", easyQuestion())
		res, err := f.uc.Grade(ctx, graderToken(t), gradeReq(`["a1"]`))
		require.NoError(t, err)
		assert.Equal(t, "Ollama API returned status: 502", res.Results[0].Error)
	})
}

func gradePrompt(t *testing.T, menu, rules string, q *entity.AssessmentQuestion) string {
	t.Helper()
	oll := &fakeOllama{text: `{"grades":{},"total_overall_score":1}`}
	f := newGrade(t, usecase.AIClients{Ollama: oll})
	asisten(f)
	f.expectAttempt(submitted("a1", menu), rules, *q)
	f.assessments.On("SaveGrade", mock.Anything, "a1", mock.Anything, json.Number("1"), gradeNow).Return(nil)
	_, err := f.uc.Grade(context.Background(), graderToken(t), gradeReq(`["a1"]`))
	require.NoError(t, err)
	require.Len(t, oll.prompts, 1)
	return oll.prompts[0]
}

func TestGradePromptVerbatim(t *testing.T) {
	q := easyQuestion()
	p := gradePrompt(t, "pre_test", "", &q)
	want := "Anda adalah Asisten Praktikum AI. Anda ditugaskan menilai pengerjaan ujian mahasiswa bernama \"Budi\" (NIM: 202711001).\n" +
		"Tipe Asesmen: PRE_TEST\n" +
		"Evaluasi seluruh soal berikut dan berikan skor sesuai kriteria penilaian.\n\n" +
		"Berikut adalah daftar soal dan jawaban draf mahasiswa:\n" +
		"======================================================\n" +
		"SOAL 1 (ID Soal: q1):\n" +
		"- Judul Soal: Soal A\n" +
		"- Instruksi Soal: Kerjakan\n" +
		"- Tipe Soal: essay\n" +
		"- Kriteria Kesulitan: easy\n" +
		"- Modul Terkait: Modul N/A\n" +
		"- Kode Awal / Template: \n```\n\n```\n" +
		"- Kunci Jawaban Referensi / Solusi:\n```\n\n```\n\n" +
		"JAWABAN MAHASISWA (SOAL 1):\n" +
		"- Teks Jawaban / Penjelasan: \"jawab\"\n" +
		"- Kode yang Dikirim:\n```\nprint(1)\n```\n" +
		"- Hasil Output Sandbox (Stdout): \"1\"\n" +
		"- Error Sandbox (Stderr): \"\"\n\n" +
		"KRITERIA SCORING KHUSUS UNTUK SOAL INI:\n" +
		"- Pilihan Ganda / Jawaban Singkat Easy: Jawaban Benar = 20 poin, Jawaban Salah = 8 poin, Kosong = 0 poin.\n" +
		"======================================================\n\n" +
		"TUGAS ANDA:\n"
	assert.True(t, strings.HasPrefix(p, want), "prompt:\n%s", p)
	assert.True(t, strings.HasSuffix(p, "  \"overall_feedback\": \"Komentar umum pengerjaan asesmen mahasiswa ini...\"\n}"))
}

func TestGradePromptCriteriaOrder(t *testing.T) {
	q := easyQuestion()
	rules := `{"difficulties":{"easy":{"criteria":{"zeta_benar_sekali":20,"alpha_salah":8.5,"10":3,"2":1}}}}`
	p := gradePrompt(t, "post_test", rules, &q)
	assert.Contains(t, p, "KRITERIA SCORING KHUSUS UNTUK SOAL INI:\n- EASY: 2: 1 poin, 10: 3 poin, zeta benar sekali: 20 poin, alpha salah: 8.5 poin\n=====")

	q.Type = strp("coding")
	p = gradePrompt(t, "ujian_praktik", `{"soal_1":{"criteria":{"jalan_tanpa_error":7,"sesuai":25}}}`, &q)
	assert.Contains(t, p, "- Ujian Praktik SOAL 1: jalan tanpa error: 7 poin, sesuai: 25 poin (Maks 0 poin).\n")

	p = gradePrompt(t, "program_keterampilan", `{"criteria":[{"label":"Berjalan","nilai":30},{"label":"Sintaks","nilai":35}]}`, &q)
	assert.Contains(t, p, "- Program Keterampilan: Berjalan: 30 poin, Sintaks: 35 poin (Total 85).\n")
}

func TestGradeConcurrencyLimitAndOrder(t *testing.T) {
	var inflight, peak atomic.Int32
	release := make(chan struct{})
	g := &fakeGemini{text: `{"grades":{},"total_overall_score":10}`, hook: func() {
		n := inflight.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		<-release
		inflight.Add(-1)
	}}
	f := newGrade(t, usecase.AIClients{Gemini: g, GeminiKeys: []string{"k1"}})
	asisten(f)
	ids := []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7"}
	f.assessments.On("Attempt", mock.Anything, mock.Anything).Return(func(_ context.Context, id string) (*entity.AssessmentAttempt, error) {
		return submitted(id, "pre_test"), nil
	})
	f.users.On("NamaByNIM", mock.Anything, "202711001").Return("Budi", nil)
	f.assessments.On("GradingRules", mock.Anything, "pre_test").Return(nil, gorm.ErrRecordNotFound)
	f.assessments.On("Questions", mock.Anything, []string{"q1"}).Return([]entity.AssessmentQuestion{easyQuestion()}, nil)
	f.assessments.On("SaveGrade", mock.Anything, mock.Anything, json.RawMessage(`{}`), json.Number("10"), gradeNow).Return(nil)

	raw, _ := json.Marshal(ids)
	done := make(chan *dto.GradeResponse)
	go func() {
		res, err := f.uc.Grade(context.Background(), graderToken(t), gradeReq(string(raw)))
		assert.NoError(t, err)
		done <- res
	}()

	require.Eventually(t, func() bool { return inflight.Load() == 3 }, 5*time.Second, time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(3), inflight.Load(), "attempt keempat tidak boleh mulai sebelum slot kosong")
	close(release)

	res := <-done
	assert.Equal(t, int32(3), peak.Load())
	assert.Equal(t, 7, res.GradedCount)
	for i, r := range res.Results {
		assert.Equal(t, ids[i], r.AttemptID)
	}
}
