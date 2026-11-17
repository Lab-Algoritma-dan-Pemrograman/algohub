package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type AchievementRepository interface {
	List(ctx context.Context, requirementType string) ([]entity.Achievement, error)
	UnlockedIDs(ctx context.Context, nim string) ([]string, error)
	Insert(ctx context.Context, nim, achievementID string, unlockedAt time.Time) error
	DeleteByNIM(ctx context.Context, nim string) (int64, error)
}

type achievementRepository struct{ db *gorm.DB }

func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &achievementRepository{db: db}
}

// Katalog pencapaian adalah tabel seed kecil yang dibaca utuh oleh checkAndUnlockAchievements, jadi tidak dipaginasi.
func (r *achievementRepository) List(ctx context.Context, requirementType string) ([]entity.Achievement, error) {
	q := r.db.WithContext(ctx)
	if requirementType != "" {
		q = q.Where("requirement_type = ?", requirementType)
	}
	var rows []entity.Achievement
	err := q.Find(&rows).Error
	return rows, err
}

func (r *achievementRepository) UnlockedIDs(ctx context.Context, nim string) ([]string, error) {
	ids := []string{}
	err := r.db.WithContext(ctx).Model(&entity.UnlockedAchievement{}).Where("nim = ?", nim).Pluck("achievement_id", &ids).Error
	return ids, err
}

func (r *achievementRepository) Insert(ctx context.Context, nim, achievementID string, unlockedAt time.Time) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO unlocked_achievements (nim, achievement_id, unlocked_at) VALUES (?, ?, ?)", nim, achievementID, unlockedAt).Error
}

func (r *achievementRepository) DeleteByNIM(ctx context.Context, nim string) (int64, error) {
	res := r.db.WithContext(ctx).Where("nim = ?", nim).Delete(&entity.UnlockedAchievement{})
	return res.RowsAffected, res.Error
}
