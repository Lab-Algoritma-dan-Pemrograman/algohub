package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

// Kolom persis payload insert gameService.ts.
var questionColumns = []string{"language", "difficulty", "title", "code", "bug_line", "explanation"}

type QuestionFields struct {
	Language    *string
	Difficulty  *string
	Title       *string
	Code        *string
	BugLine     *int
	Explanation *string
}

func (q *QuestionFields) values() []any {
	return []any{q.Language, q.Difficulty, q.Title, q.Code, q.BugLine, q.Explanation}
}

// Baris settings tunggal game_settings memakai id 'default' (gameService.ts:254).
const gameSettingsID = "default"

type GameRepository interface {
	QuestionsByLanguage(ctx context.Context, language string) ([]entity.GameQuestion, error)
	CountQuestions(ctx context.Context) (int64, error)
	InsertQuestions(ctx context.Context, rows []QuestionFields) error
	ReplaceAll(ctx context.Context, rows []QuestionFields) error
	InsertQuestion(ctx context.Context, q *QuestionFields) (string, error)
	UpdateQuestion(ctx context.Context, id string, cols map[string]any) (int64, error)
	DeleteQuestion(ctx context.Context, id string) (int64, error)
	Settings(ctx context.Context) (*entity.GameSettings, error)
	UpsertSettings(ctx context.Context, cols map[string]any) error
	CountPlaysSince(ctx context.Context, nim, gameType string, since time.Time) (int64, error)
	InsertHistory(ctx context.Context, nim, gameType string, xpEarned int, playedAt time.Time) error
}

type gameRepository struct{ db *gorm.DB }

func NewGameRepository(db *gorm.DB) GameRepository { return &gameRepository{db: db} }

// Tanpa urutan karena getGameQuestions mengacak seluruh hasil di aplikasi.
func (r *gameRepository) QuestionsByLanguage(ctx context.Context, language string) ([]entity.GameQuestion, error) {
	var rows []entity.GameQuestion
	err := r.db.WithContext(ctx).Where("language = ?", language).Find(&rows).Error
	return rows, err
}

func (r *gameRepository) CountQuestions(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.GameQuestion{}).Count(&n).Error
	return n, err
}

func questionRows(rows []QuestionFields) [][]any {
	out := make([][]any, len(rows))
	for i := range rows {
		out[i] = rows[i].values()
	}
	return out
}

func (r *gameRepository) InsertQuestions(ctx context.Context, rows []QuestionFields) error {
	return insertRows(r.db.WithContext(ctx), "game_questions", questionColumns, questionRows(rows))
}

// Dulu hapus lalu insert tanpa transaksi (gameService.ts:176-195); kini insert yang gagal tidak meninggalkan bank soal kosong.
func (r *gameRepository) ReplaceAll(ctx context.Context, rows []QuestionFields) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM game_questions").Error; err != nil {
			return err
		}
		return insertRows(tx, "game_questions", questionColumns, questionRows(rows))
	})
}

func (r *gameRepository) InsertQuestion(ctx context.Context, q *QuestionFields) (string, error) {
	cols := map[string]any{}
	for i, v := range q.values() {
		cols[questionColumns[i]] = v
	}
	return insertReturningID(ctx, r.db, "game_questions", cols)
}

func (r *gameRepository) UpdateQuestion(ctx context.Context, id string, cols map[string]any) (int64, error) {
	return updateByID(ctx, r.db, "game_questions", id, cols)
}

func (r *gameRepository) DeleteQuestion(ctx context.Context, id string) (int64, error) {
	return deleteByID(ctx, r.db, "game_questions", id)
}

// Hasil nil berarti tidak tepat satu baris, seperti .single() yang gagal di getGameSettings.
func (r *gameRepository) Settings(ctx context.Context) (*entity.GameSettings, error) {
	var rows []entity.GameSettings
	if err := r.db.WithContext(ctx).Where("id = ?", gameSettingsID).Limit(2).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, nil
	}
	return &rows[0], nil
}

// Setara .upsert({id: 'default', ...formatted}): kolom yang tidak dikirim tidak disentuh saat konflik.
func (r *gameRepository) UpsertSettings(ctx context.Context, cols map[string]any) error {
	all := map[string]any{"id": gameSettingsID}
	for k, v := range cols {
		all[k] = v
	}
	names, params, args := assignments(all)
	var sets []string
	for _, n := range names {
		if n != "id" {
			sets = append(sets, n+" = EXCLUDED."+n)
		}
	}
	conflict := "DO NOTHING"
	if len(sets) > 0 {
		conflict = "DO UPDATE SET " + strings.Join(sets, ", ")
	}
	sql := "INSERT INTO game_settings (" + strings.Join(names, ", ") + ") VALUES (" + strings.Join(params, ", ") + ") ON CONFLICT (id) " + conflict
	return r.db.WithContext(ctx).Exec(sql, args).Error
}

func (r *gameRepository) CountPlaysSince(ctx context.Context, nim, gameType string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.GameHistory{}).Where("nim = ? AND game_type = ? AND played_at >= ?", nim, gameType, since).Count(&n).Error
	return n, err
}

func (r *gameRepository) InsertHistory(ctx context.Context, nim, gameType string, xpEarned int, playedAt time.Time) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO game_history (nim, game_type, xp_earned, played_at) VALUES (?, ?, ?, ?)", nim, gameType, xpEarned, playedAt).Error
}
