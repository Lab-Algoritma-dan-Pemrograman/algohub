package usecase

import (
	"cmp"
	"context"
	"math/rand/v2"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

const (
	gameBugHunt          = "bug_hunt"
	msgWeeklyLimit       = "WEEKLY_LIMIT_REACHED"
	defaultQuestionCount = 5
)

type GameUsecase struct {
	games   repository.GameRepository
	users   repository.UserRepository
	guard   userGuard
	now     func() time.Time
	shuffle func(n int, swap func(i, j int))
}

// shuffle nil berarti acak sungguhan; test menyuntikkan urutan tetap.
func NewGameUsecase(games repository.GameRepository, users repository.UserRepository, now func() time.Time, shuffle func(n int, swap func(i, j int))) *GameUsecase {
	if shuffle == nil {
		shuffle = rand.Shuffle
	}
	return &GameUsecase{games: games, users: users, guard: userGuard{users: users}, now: now, shuffle: shuffle}
}

// Disalin dari getGameQuestions: semua soal satu bahasa diacak lalu dipotong.
func (uc *GameUsecase) Questions(ctx context.Context, p *dto.Principal, language string, limit int) ([]dto.GameQuestion, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	rows, err := uc.games.QuestionsByLanguage(ctx, language)
	if err != nil {
		return nil, err
	}
	uc.shuffle(len(rows), func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	out := make([]dto.GameQuestion, 0, min(limit, len(rows)))
	for i := range rows[:min(limit, len(rows))] {
		q := &rows[i]
		out = append(out, dto.GameQuestion{ID: q.ID, Language: q.Language, Difficulty: q.Difficulty, Title: q.Title, Code: q.Code, BugLine: q.BugLine, Explanation: q.Explanation})
	}
	return out, nil
}

func questionFields(in []dto.GameQuestionInput) []repository.QuestionFields {
	out := make([]repository.QuestionFields, len(in))
	for i, q := range in {
		out[i] = repository.QuestionFields{Language: q.Language, Difficulty: q.Difficulty, Title: q.Title, Code: q.Code, BugLine: q.BugLine, Explanation: q.Explanation}
	}
	return out
}

// Disalin dari seedInitialQuestions: hanya mengisi bila bank soal kosong.
func (uc *GameUsecase) Seed(ctx context.Context, p *dto.Principal, req *dto.GameQuestionsRequest) (*dto.SeedResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	n, err := uc.games.CountQuestions(ctx)
	if err != nil {
		return nil, err
	}
	if n != 0 {
		return &dto.SeedResponse{}, nil
	}
	if len(req.Questions) > 0 && !IsAdminOrKordas(p.Role) {
		return nil, RLSInsert("game_questions")
	}
	if err := uc.games.InsertQuestions(ctx, questionFields(req.Questions)); err != nil {
		return nil, err
	}
	return &dto.SeedResponse{Seeded: true}, nil
}

// forceResetGameQuestions: DELETE oleh selain admin/kordas terkena USING (0 baris) dan INSERT-nya ditolak, jadi tidak ada yang berubah.
func (uc *GameUsecase) ForceReset(ctx context.Context, p *dto.Principal, req *dto.GameQuestionsRequest) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		if len(req.Questions) == 0 {
			return nil
		}
		return RLSInsert("game_questions")
	}
	return uc.games.ReplaceAll(ctx, questionFields(req.Questions))
}

func (uc *GameUsecase) Add(ctx context.Context, p *dto.Principal, req *dto.GameQuestionInput) (*dto.IDResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil, RLSInsert("game_questions")
	}
	id, err := uc.games.InsertQuestion(ctx, &questionFields([]dto.GameQuestionInput{*req})[0])
	if err != nil {
		return nil, err
	}
	return &dto.IDResponse{ID: id}, nil
}

// Disalin dari updateGameQuestion: string kosong dilewati, bugLine ditulis bila dikirim (termasuk null).
func (uc *GameUsecase) Update(ctx context.Context, p *dto.Principal, id string, req *dto.GameQuestionUpdate) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	cols := filledColumns(map[string]string{"language": req.Language, "difficulty": req.Difficulty, "title": req.Title, "code": req.Code, "explanation": req.Explanation})
	if req.BugLine.Set {
		cols["bug_line"] = req.BugLine.Value
	}
	_, err := uc.games.UpdateQuestion(ctx, id, cols)
	return err
}

func filledColumns(fields map[string]string) map[string]any {
	cols := map[string]any{}
	for k, v := range fields {
		if v != "" {
			cols[k] = v
		}
	}
	return cols
}

func (uc *GameUsecase) Delete(ctx context.Context, p *dto.Principal, id string) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	_, err := uc.games.DeleteQuestion(ctx, id)
	return err
}

func (uc *GameUsecase) Settings(ctx context.Context, p *dto.Principal) (*dto.GameSettings, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	return uc.settings(ctx), nil
}

func ptr[T any](v T) *T { return &v }

// Disalin dari getGameSettings: baris tidak ada maupun error memakai nilai bawaan.
func (uc *GameUsecase) settings(ctx context.Context) *dto.GameSettings {
	s, err := uc.games.Settings(ctx)
	if err != nil || s == nil {
		return &dto.GameSettings{BugHuntActive: ptr(true), BugHuntCActive: ptr(true), BugHuntPythonActive: ptr(true), BugHuntWeeklyLimit: ptr(3), BugHuntQuestionCount: defaultQuestionCount}
	}
	return &dto.GameSettings{
		BugHuntActive:        s.BugHuntActive,
		BugHuntCActive:       s.BugHuntCActive,
		BugHuntPythonActive:  s.BugHuntPythonActive,
		BugHuntWeeklyLimit:   s.BugHuntWeeklyLimit,
		BugHuntQuestionCount: defaultQuestionCount,
	}
}

// null dibandingkan sebagai 0, sama dengan `settings.bugHuntWeeklyLimit > 0` di JavaScript.
func weeklyLimit(s *dto.GameSettings) int {
	if s.BugHuntWeeklyLimit == nil {
		return 0
	}
	return *s.BugHuntWeeklyLimit
}

func (uc *GameUsecase) UpdateSettings(ctx context.Context, p *dto.Principal, req *dto.GameSettingsUpdate) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return RLSInsert("game_settings")
	}
	cols := map[string]any{}
	if req.BugHuntActive.Set {
		cols["bug_hunt_active"] = req.BugHuntActive.Value
	}
	if req.BugHuntCActive.Set {
		cols["bug_hunt_c_active"] = req.BugHuntCActive.Value
	}
	if req.BugHuntPythonActive.Set {
		cols["bug_hunt_python_active"] = req.BugHuntPythonActive.Value
	}
	if req.BugHuntWeeklyLimit.Set {
		cols["bug_hunt_weekly_limit"] = req.BugHuntWeeklyLimit.Value
	}
	return uc.games.UpsertSettings(ctx, cols)
}

// Disalin dari getPlaysThisWeek; riwayat orang lain tersaring RLS untuk non-staf (R-game_history-SELECT-1).
func (uc *GameUsecase) playsThisWeek(ctx context.Context, p *dto.Principal, nim string) (int64, error) {
	if !ownsOrStaff(p, nim) {
		return 0, nil
	}
	return uc.games.CountPlaysSince(ctx, nim, gameBugHunt, jakartaWeekOf(uc.now()).start)
}

// Disalin dari canPlayBugHunt.
func (uc *GameUsecase) PlaysMe(ctx context.Context, p *dto.Principal) (*dto.BugHuntAccess, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	s := uc.settings(ctx)
	if s.BugHuntActive == nil || !*s.BugHuntActive {
		return &dto.BugHuntAccess{}, nil
	}
	limit := weeklyLimit(s)
	if limit <= 0 {
		return &dto.BugHuntAccess{Allowed: true}, nil
	}
	used, err := uc.playsThisWeek(ctx, p, p.NIM)
	if err != nil {
		return nil, err
	}
	return &dto.BugHuntAccess{Allowed: used < int64(limit), PlaysUsed: used, Limit: limit}, nil
}

// Disalin dari saveGameHistory: game_history dulu supaya check_user_xp_level sudah melihat baris baru, lalu users.
func (uc *GameUsecase) History(ctx context.Context, p *dto.Principal, req *dto.GameHistoryRequest) (*dto.GameHistoryResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	nim := p.NIM
	if req.GameType == gameBugHunt {
		if limit := weeklyLimit(uc.settings(ctx)); limit > 0 {
			used, err := uc.playsThisWeek(ctx, p, nim)
			if err != nil {
				return nil, err
			}
			if used >= int64(limit) {
				return nil, BadRequest(msgWeeklyLimit)
			}
		}
	}

	now := uc.now().UTC()
	if nim != p.AuthNIM {
		return nil, RLSInsert("game_history")
	}
	if err := uc.games.InsertHistory(ctx, nim, req.GameType, req.XPEarned, now); err != nil {
		return nil, MapDBError(err)
	}

	u, err := loadProfile(ctx, uc.users, nim)
	if err != nil {
		return nil, err
	}
	newXP := u.XP + req.XPEarned
	newLevel := calculateLevel(newXP)
	streak, _ := calculateStreak(u.LastActive, u.Streak, now)
	cols := map[string]any{"xp": newXP, "level": newLevel, "streak": streak, "last_active": now}
	if _, err := uc.guard.update(ctx, p, nim, cols); err != nil {
		return nil, err
	}
	return &dto.GameHistoryResponse{XP: newXP, Level: newLevel, Streak: streak, LastActive: now, LeveledUp: newLevel > cmp.Or(u.Level, 1)}, nil
}
