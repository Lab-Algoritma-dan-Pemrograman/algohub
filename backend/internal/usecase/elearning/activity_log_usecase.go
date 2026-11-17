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

// monitoringService.addAuditLog tidak memeriksa error, jadi kegagalan mencatat tidak menggagalkan aksi admin.
func (uc *ActivityLogUsecase) audit(ctx context.Context, p *dto.Principal, eventType, details string) {
	if err := uc.insert(ctx, p, p.NIM, p.Nama, eventType, details); err != nil {
		log.Printf("elearning: audit %s: %v", eventType, err)
	}
}
