package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

// Join mengikuti embed PostgREST lama `schedule:schedules(*), user:users(*)` lewat FK schedule_id dan user_id.
const todayAssignmentsSQL = "SELECT sa.user_id, s.start_time::text, u.role::text, sa.task_role::text, sa.activity_name::text " +
	"FROM public.schedule_assignments sa " +
	"LEFT JOIN public.schedules s ON s.id = sa.schedule_id " +
	"LEFT JOIN public.users u ON u.id = sa.user_id " +
	"WHERE sa.activity_date = ? AND (sa.status IS NULL OR sa.status = 'aktif')"

type ShiftRepository interface {
	TodayAssignments(ctx context.Context, date string) ([]entity.ShiftAssignment, error)
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) TodayAssignments(ctx context.Context, date string) ([]entity.ShiftAssignment, error) {
	rows, err := r.db.WithContext(ctx).Raw(todayAssignmentsSQL, date).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.ShiftAssignment
	for rows.Next() {
		var (
			userID                              sql.NullInt64
			startTime, role, taskRole, activity sql.NullString
		)
		if err := rows.Scan(&userID, &startTime, &role, &taskRole, &activity); err != nil {
			return nil, err
		}
		a := entity.ShiftAssignment{
			StartTime:    startTime.String,
			UserRole:     role.String,
			TaskRole:     taskRole.String,
			ActivityName: activity.String,
		}
		if userID.Valid {
			a.UserID = &userID.Int64
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
