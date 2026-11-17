package usecase

import (
	"bytes"
	"cmp"
	"context"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
	"strings"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
)

type GeminiGenerator interface {
	GenerateParts(ctx context.Context, apiKey, model string, parts []gemini.Part, cfg gemini.GenerationConfig) (string, error)
}

type OllamaChatter interface {
	NativeChatJSON(ctx context.Context, prompt string) (string, error)
}

type AIClients struct {
	Gemini     GeminiGenerator
	GeminiKeys []string
	// Nil bila ELEARNING_OLLAMA_URL kosong; api/grade.ts lalu memakai Gemini.
	Ollama OllamaChatter
	// Nil = acak; diinjeksi supaya urutan rotasi kunci bisa diuji.
	Shuffle func([]string)
}

// Rotasi maksimal min(n, 3) kunci dalam urutan acak, sama dengan api/ai.ts dan api/grade.ts.
func (a *AIClients) rotation() []string {
	keys := slices.Clone(a.GeminiKeys)
	if a.Shuffle != nil {
		a.Shuffle(keys)
	} else {
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	}
	return keys[:min(len(keys), 3)]
}

var aiModels = []string{"gemini-3-flash-preview", "gemini-2.5-flash"}

type AIUsecase struct {
	auth  *AuthUsecase
	users repository.UserRepository
	ai    AIClients
}

func NewAIUsecase(auth *AuthUsecase, users repository.UserRepository, ai AIClients) *AIUsecase {
	return &AIUsecase{auth: auth, users: users, ai: ai}
}

func (uc *AIUsecase) Generate(ctx context.Context, token string, req *dto.AIRequest) (*dto.AIResponse, error) {
	// Kunci dicek sebelum token, sama dengan urutan api/ai.ts.
	if len(uc.ai.GeminiKeys) == 0 {
		return nil, newError(http.StatusInternalServerError, "AI Keys not configured on server")
	}
	if token == "" {
		return nil, Unauthorized("Unauthorized: Missing token")
	}
	claims, err := uc.auth.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}
	nim, ok := claims["nim"]
	if !ok || nim == nil {
		return nil, Forbidden("Unauthorized: User not found in database")
	}
	role, err := uc.users.RoleByNIM(ctx, jsString(nim))
	if err != nil {
		return nil, Forbidden("Unauthorized: User not found in database")
	}
	// Dibandingkan dengan users.role mentah seperti kode lama, jadi nilai kanonik "koordinator" ikut ditolak.
	if role != RoleAdmin && role != RoleKordas {
		return nil, Forbidden("Akses Ditolak: Fitur AI ini hanya tersedia untuk Admin atau Koordinator.")
	}

	model := "gemini-3-flash-preview"
	if slices.Contains(aiModels, req.Model) {
		model = req.Model
	}
	parts := []gemini.Part{{Text: req.Prompt}}
	if req.FileData != "" {
		mime := "application/pdf"
		if req.FileMimeType != nil {
			mime = *req.FileMimeType
		}
		parts = append(parts, gemini.Part{InlineData: &gemini.InlineData{MimeType: mime, Data: req.FileData}})
	}
	cfg := gemini.GenerationConfig{ResponseMimeType: cmp.Or(req.ResponseMimeType, "text/plain")}
	if s := bytes.TrimSpace(req.ResponseSchema); len(s) > 0 && !bytes.Equal(s, []byte("null")) {
		cfg.ResponseSchema = s
	}

	var lastErr error
	for i, key := range uc.ai.rotation() {
		text, err := uc.ai.Gemini.GenerateParts(ctx, key, model, parts, cfg)
		if err == nil {
			return &dto.AIResponse{Text: text}, nil
		}
		lastErr = err
		log.Printf("elearning ai: percobaan %d gagal: %v", i+1, err)
		// Filter safety atau model tidak ada tidak akan berhasil dengan kunci lain.
		msg := err.Error()
		if strings.Contains(msg, "safety") || strings.Contains(msg, "not found") || strings.Contains(msg, "404") {
			break
		}
	}
	return nil, newError(http.StatusInternalServerError, "Gagal memproses AI. Detail: "+lastErr.Error())
}
