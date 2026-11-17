package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type PlaygroundRepository interface {
	List(ctx context.Context, language string, page, limit int) ([]entity.PlaygroundExample, int64, error)
	Insert(ctx context.Context, cols map[string]any) (string, error)
	Update(ctx context.Context, id string, cols map[string]any) (int64, error)
	Delete(ctx context.Context, id string) (int64, error)
}

type playgroundRepository struct{ db *gorm.DB }

func NewPlaygroundRepository(db *gorm.DB) PlaygroundRepository { return &playgroundRepository{db: db} }

// Urutan sort_order dari getPlaygroundExamples; id menjadi pemecah seri supaya halaman stabil.
func (r *playgroundRepository) List(ctx context.Context, language string, page, limit int) ([]entity.PlaygroundExample, int64, error) {
	filter := noFilter
	if language != "" {
		filter = func(db *gorm.DB) *gorm.DB { return db.Where("language = ?", language) }
	}
	return findPage[entity.PlaygroundExample](ctx, filter, r.db, "sort_order, id", page, limit)
}

func (r *playgroundRepository) Insert(ctx context.Context, cols map[string]any) (string, error) {
	return insertReturningID(ctx, r.db, "playground_examples", cols)
}

func (r *playgroundRepository) Update(ctx context.Context, id string, cols map[string]any) (int64, error) {
	return updateByID(ctx, r.db, "playground_examples", id, cols)
}

func (r *playgroundRepository) Delete(ctx context.Context, id string) (int64, error) {
	return deleteByID(ctx, r.db, "playground_examples", id)
}
