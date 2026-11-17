package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
)

type geminiCall struct {
	key, model string
	parts      []gemini.Part
	cfg        gemini.GenerationConfig
}

// Fake tanpa jaringan: error per kunci, selain itu mengembalikan text.
type fakeGemini struct {
	mu    sync.Mutex
	calls []geminiCall
	errs  map[string]error
	text  string
	hook  func()
}

func (f *fakeGemini) GenerateParts(_ context.Context, key, model string, parts []gemini.Part, cfg gemini.GenerationConfig) (string, error) {
	f.mu.Lock()
	f.calls = append(f.calls, geminiCall{key: key, model: model, parts: parts, cfg: cfg})
	err := f.errs[key]
	f.mu.Unlock()
	if f.hook != nil {
		f.hook()
	}
	if err != nil {
		return "", err
	}
	return f.text, nil
}

func (f *fakeGemini) keys() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = c.key
	}
	return out
}

func reverseKeys(k []string) { slices.Reverse(k) }

func supabaseToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	return sign(t, supabaseSecret, jwt.SigningMethodHS256, claims)
}

func newAIUsecase(t *testing.T, keys []string, g *fakeGemini) (*usecase.AIUsecase, *mocks.UserRepository) {
	t.Helper()
	users := mocks.NewUserRepository(t)
	ai := usecase.AIClients{Gemini: g, GeminiKeys: keys, Shuffle: reverseKeys}
	return usecase.NewAIUsecase(newAuth(testConfig(), nil), users, ai), users
}

func TestAIChecksKeysBeforeToken(t *testing.T) {
	uc, _ := newAIUsecase(t, nil, &fakeGemini{})
	_, err := uc.Generate(context.Background(), "", &dto.AIRequest{})
	assertUCError(t, err, http.StatusInternalServerError, "AI Keys not configured on server")
}

func TestAITokenAndRoleChecks(t *testing.T) {
	ctx := context.Background()
	t.Run("tanpa token", func(t *testing.T) {
		uc, _ := newAIUsecase(t, []string{"k1"}, &fakeGemini{})
		_, err := uc.Generate(ctx, "", &dto.AIRequest{})
		assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Missing token")
	})
	t.Run("token rusak", func(t *testing.T) {
		uc, _ := newAIUsecase(t, []string{"k1"}, &fakeGemini{})
		_, err := uc.Generate(ctx, "rusak", &dto.AIRequest{})
		assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")
	})
	t.Run("claim nim saja, bukan sub", func(t *testing.T) {
		uc, _ := newAIUsecase(t, []string{"k1"}, &fakeGemini{})
		_, err := uc.Generate(ctx, supabaseToken(t, jwt.MapClaims{"sub": "202711001"}), &dto.AIRequest{})
		assertUCError(t, err, http.StatusForbidden, "Unauthorized: User not found in database")
	})
	t.Run("user tidak ada", func(t *testing.T) {
		uc, users := newAIUsecase(t, []string{"k1"}, &fakeGemini{})
		users.On("RoleByNIM", mock.Anything, "202711001").Return("", gorm.ErrRecordNotFound)
		_, err := uc.Generate(ctx, supabaseToken(t, jwt.MapClaims{"nim": "202711001"}), &dto.AIRequest{})
		assertUCError(t, err, http.StatusForbidden, "Unauthorized: User not found in database")
	})
	for _, role := range []string{"koordinator", "asisten", "mahasiswa", ""} {
		t.Run("role mentah "+role, func(t *testing.T) {
			g := &fakeGemini{}
			uc, users := newAIUsecase(t, []string{"k1"}, g)
			users.On("RoleByNIM", mock.Anything, "202711001").Return(role, nil)
			_, err := uc.Generate(ctx, supabaseToken(t, jwt.MapClaims{"nim": "202711001", "user_role": "kordas"}), &dto.AIRequest{})
			assertUCError(t, err, http.StatusForbidden, "Akses Ditolak: Fitur AI ini hanya tersedia untuk Admin atau Koordinator.")
			assert.Empty(t, g.calls)
		})
	}
}

func TestAIRequestShape(t *testing.T) {
	g := &fakeGemini{text: "jawaban"}
	uc, users := newAIUsecase(t, []string{"k1"}, g)
	users.On("RoleByNIM", mock.Anything, "202711001").Return("kordas", nil)
	tok := supabaseToken(t, jwt.MapClaims{"nim": "202711001"})

	res, err := uc.Generate(context.Background(), tok, &dto.AIRequest{Prompt: "halo", Model: "gemini-2.5-flash", FileData: "QUJD", ResponseSchema: json.RawMessage(`{"type":"OBJECT"}`)})
	require.NoError(t, err)
	assert.Equal(t, "jawaban", res.Text)
	c := g.calls[0]
	assert.Equal(t, "gemini-2.5-flash", c.model)
	assert.Equal(t, []gemini.Part{{Text: "halo"}, {InlineData: &gemini.InlineData{MimeType: "application/pdf", Data: "QUJD"}}}, c.parts)
	assert.Equal(t, "text/plain", c.cfg.ResponseMimeType)
	assert.JSONEq(t, `{"type":"OBJECT"}`, string(c.cfg.ResponseSchema))

	png := "image/png"
	_, err = uc.Generate(context.Background(), tok, &dto.AIRequest{Prompt: "x", Model: "gpt-4", ResponseMimeType: "application/json", FileData: "AA", FileMimeType: &png, ResponseSchema: json.RawMessage("null")})
	require.NoError(t, err)
	c = g.calls[1]
	assert.Equal(t, "gemini-3-flash-preview", c.model)
	assert.Equal(t, "image/png", c.parts[1].InlineData.MimeType)
	assert.Equal(t, "application/json", c.cfg.ResponseMimeType)
	assert.Nil(t, c.cfg.ResponseSchema)
}

func TestAIKeyRotation(t *testing.T) {
	tok := func(t *testing.T) string { return supabaseToken(t, jwt.MapClaims{"nim": "202711001"}) }
	setup := func(t *testing.T, g *fakeGemini) *usecase.AIUsecase {
		uc, users := newAIUsecase(t, []string{"k1", "k2", "k3", "k4"}, g)
		users.On("RoleByNIM", mock.Anything, "202711001").Return("admin", nil)
		return uc
	}

	t.Run("pindah kunci sampai berhasil", func(t *testing.T) {
		g := &fakeGemini{text: "ok", errs: map[string]error{"k4": errors.New("gemini: status 429: quota"), "k3": errors.New("gemini: status 500: internal")}}
		res, err := setup(t, g).Generate(context.Background(), tok(t), &dto.AIRequest{Prompt: "p"})
		require.NoError(t, err)
		assert.Equal(t, "ok", res.Text)
		assert.Equal(t, []string{"k4", "k3", "k2"}, g.keys())
	})
	t.Run("maksimal tiga kunci", func(t *testing.T) {
		quota := errors.New("gemini: status 429: quota")
		g := &fakeGemini{errs: map[string]error{"k1": quota, "k2": quota, "k3": quota, "k4": errors.New("gemini: status 503: sibuk")}}
		_, err := setup(t, g).Generate(context.Background(), tok(t), &dto.AIRequest{Prompt: "p"})
		assertUCError(t, err, http.StatusInternalServerError, "Gagal memproses AI. Detail: gemini: status 429: quota")
		assert.Equal(t, []string{"k4", "k3", "k2"}, g.keys())
	})
	for _, msg := range []string{"gemini: status 400: blocked by safety settings", "gemini: status 400: model is not found", "gemini: status 404: x"} {
		t.Run("berhenti: "+msg, func(t *testing.T) {
			g := &fakeGemini{errs: map[string]error{"k4": errors.New(msg)}}
			_, err := setup(t, g).Generate(context.Background(), tok(t), &dto.AIRequest{Prompt: "p"})
			assertUCError(t, err, http.StatusInternalServerError, "Gagal memproses AI. Detail: "+msg)
			assert.Equal(t, []string{"k4"}, g.keys())
		})
	}
}
