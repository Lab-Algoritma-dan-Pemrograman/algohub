package usecase

import (
	"cmp"
	"context"
	"log"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Ambang cleanupStaleSessions(2) di StudentMonitoring.tsx.
const DefaultStaleMinutes = 2

type SessionUsecase struct {
	sessions repository.SessionRepository
	users    repository.UserRepository
	progress repository.ProgressRepository
	learning *ProgressUsecase
	now      func() time.Time
}

func NewSessionUsecase(sessions repository.SessionRepository, users repository.UserRepository, progress repository.ProgressRepository, learning *ProgressUsecase, now func() time.Time) *SessionUsecase {
	return &SessionUsecase{sessions: sessions, users: users, progress: progress, learning: learning, now: now}
}

// Heartbeat Layout.tsx:194-247. Upsert active_sessions oleh praktikan ditolak RLS karena baris baru tidak lolos policy
// SELECT staf (R-active_sessions-INSERT-1, tidak pasti); Layout menelan error itu lalu tetap memperbarui users, jadi
// penolakan di sini juga tidak menggagalkan streak dan waktu belajar.
func (uc *SessionUsecase) Heartbeat(ctx context.Context, p *dto.Principal, req *dto.HeartbeatRequest) (*dto.HeartbeatResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	res := &dto.HeartbeatResponse{}
	if IsStaff(p.Role) {
		now := uc.now().UTC()
		s := &entity.ActiveSession{
			NIM:             p.AuthNIM,
			Nama:            cmp.Or(req.Nama, p.Nama, "Anonymous"),
			Kelas:           cmp.Or(req.Kelas, p.Kelas, "Unknown"),
			LastHeartbeat:   &now,
			CurrentActivity: req.CurrentActivity,
			AttemptID:       req.AttemptID.Value,
		}
		if err := uc.sessions.Upsert(ctx, s, req.AttemptID.Set); err != nil {
			log.Printf("elearning: heartbeat %s: %v", p.AuthNIM, err)
		} else {
			res.SessionSaved = true
		}
	}
	if req.TouchStreak {
		streak, err := uc.learning.Streak(ctx, p)
		if err != nil {
			return nil, err
		}
		res.Streak = streak
	}
	if req.StudySeconds != nil && *req.StudySeconds > 0 {
		total, err := uc.learning.StudyTime(ctx, p, &dto.StudyTimeRequest{Seconds: *req.StudySeconds})
		if err != nil {
			return nil, err
		}
		res.StudyTime = total
	}
	return res, nil
}

// Logout Layout.tsx:69: DELETE dengan WHERE ikut disaring policy SELECT, jadi praktikan menghapus 0 baris (R-active_sessions-DELETE-1).
func (uc *SessionUsecase) End(ctx context.Context, p *dto.Principal) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsStaff(p.Role) {
		return nil
	}
	_, err := uc.sessions.DeleteByNIM(ctx, p.NIM)
	return err
}

// Disalin dari cleanupStaleSessions; praktikan tidak melihat baris mana pun sehingga 0 baris terhapus.
func (uc *SessionUsecase) CleanupStale(ctx context.Context, p *dto.Principal, thresholdMinutes int) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsStaff(p.Role) {
		return nil
	}
	_, err := uc.sessions.DeleteStale(ctx, uc.now().Add(-time.Duration(thresholdMinutes)*time.Minute))
	return err
}

// StudentMonitoring.tsx:61; policy SELECT active_sessions hanya untuk staf (R-active_sessions-SELECT-1).
func (uc *SessionUsecase) Sessions(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[dto.ActiveSessionRole], error) {
	return listPage(p, page, limit, func() ([]dto.ActiveSessionRole, int64, error) {
		if !IsStaff(p.Role) {
			return nil, 0, nil
		}
		rows, total, err := uc.sessions.ListWithRole(ctx, page, limit)
		out := make([]dto.ActiveSessionRole, len(rows))
		for i, r := range rows {
			out[i] = dto.ActiveSessionRole{NIM: r.NIM, LastHeartbeat: r.LastHeartbeat}
			if r.HasUser {
				out[i].Users = &dto.SessionUser{Role: r.Role}
			}
		}
		return out, total, err
	})
}

func (uc *SessionUsecase) Students(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[repository.MonitoringStudent], error) {
	return listPage(p, page, limit, func() ([]repository.MonitoringStudent, int64, error) {
		return uc.users.ListStudents(ctx, page, limit)
	})
}

// Progres orang lain hanya terbaca staf (R-student_progress-SELECT-1), jadi praktikan hanya mendapat kelompok miliknya.
func (uc *SessionUsecase) Progress(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[repository.NIMLessons], error) {
	return listPage(p, page, limit, func() ([]repository.NIMLessons, int64, error) {
		nim := ""
		if !IsStaff(p.Role) {
			nim = p.AuthNIM
		}
		return uc.progress.GroupedCompleted(ctx, nim, page, limit)
	})
}
