package usecase

import (
	"context"
	"log"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

type ActivityLogUsecase struct {
	logs repository.ActivityLogRepository
	now  func() time.Time
}

func NewActivityLogUsecase(logs repository.ActivityLogRepository, now func() time.Time) *ActivityLogUsecase {
	return &ActivityLogUsecase{logs: logs, now: now}
}

// nim selalu dari token supaya WITH CHECK nim = auth_nim() terpenuhi; nama boleh dari body seperti addAuditLog.
func (uc *ActivityLogUsecase) Insert(ctx context.Context, p *dto.Principal, req *dto.ActivityLogRequest) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	nama := p.Nama
	if req.Nama != nil {
		nama = *req.Nama
	}
	return uc.insert(ctx, p, p.AuthNIM, nama, req.EventType, req.Details)
}

func (uc *ActivityLogUsecase) insert(ctx context.Context, p *dto.Principal, nim, nama, eventType, details string) error {
	if nim != p.AuthNIM {
		return RLSInsert("activity_logs")
	}
	at := uc.now().UTC()
	return MapDBError(uc.logs.Insert(ctx, &entity.ActivityLog{NIM: &nim, Nama: &nama, EventType: eventType, Timestamp: &at, Details: &details}))
}

// Policy SELECT activity_logs hanya untuk staf (R-activity_logs-SELECT-1).
func (uc *ActivityLogUsecase) List(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[entity.ActivityLog], error) {
	return listPage(p, page, limit, func() ([]entity.ActivityLog, int64, error) {
		if !IsStaff(p.Role) {
			return nil, 0, nil
		}
		return uc.logs.List(ctx, page, limit)
	})
}

// Policy DELETE hanya admin/kordas; selain itu 0 baris tanpa error (R-activity_logs-DELETE-1).
func (uc *ActivityLogUsecase) Delete(ctx context.Context, p *dto.Principal, id string) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	_, err := uc.logs.Delete(ctx, id)
	return err
}

// monitoringService.addAuditLog tidak memeriksa error, jadi kegagalan mencatat tidak menggagalkan aksi admin.
func (uc *ActivityLogUsecase) audit(ctx context.Context, p *dto.Principal, eventType, details string) {
	if err := uc.insert(ctx, p, p.NIM, p.Nama, eventType, details); err != nil {
		log.Printf("elearning: audit %s: %v", eventType, err)
	}
}
