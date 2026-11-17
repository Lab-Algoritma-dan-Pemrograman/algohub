package usecase

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

type WASender interface {
	Send(ctx context.Context, provider, token, phone, message string) (string, error)
}

type WAUsecase struct {
	auth     notifyAuth
	notify   repository.NotifyRepository
	wa       WASender
	envToken string
}

func NewWAUsecase(notify repository.NotifyRepository, wa WASender, tokens *Tokens, envToken string) *WAUsecase {
	return &WAUsecase{auth: notifyAuth{tokens: tokens}, notify: notify, wa: wa, envToken: envToken}
}

var nonDigitRe = regexp.MustCompile(`\D`)

func failNotSent(status int, msg string) *Failure {
	return &Failure{Status: status, Body: map[string]any{"error": msg, "sent": false}}
}

func (u *WAUsecase) Send(ctx context.Context, fromCron bool, token string, req dto.SendWARequest) (map[string]any, *Failure) {
	if f := u.auth.authorize(fromCron, token, "Forbidden: Hanya staf yang boleh mengirim WhatsApp"); f != nil {
		return nil, f
	}
	if req.Message == "" {
		return nil, fail(http.StatusBadRequest, "Message content is required")
	}

	// Error dibaca sebagai settings kosong, sama dengan kode lama yang hanya mencatatnya.
	settings, _ := u.notify.WASettings(ctx)
	provider := settings.Provider
	if provider == "" {
		provider = "fonnte"
	}
	gatewayToken := settings.Token
	if gatewayToken == "" {
		gatewayToken = u.envToken
	}
	if settings.AutoNotifyEnabled != nil && !*settings.AutoNotifyEnabled {
		return map[string]any{"message": "WhatsApp auto-notify is disabled in settings", "sent": false}, nil
	}
	if gatewayToken == "" {
		return nil, failNotSent(http.StatusBadRequest,
			"WA Gateway API Token belum dikonfigurasi. Silakan isi API Token di menu Pengaturan.")
	}

	phone := req.TargetPhone
	if phone == "" && truthy(req.TargetUserID) {
		if id, ok := integerID(req.TargetUserID); ok {
			phone, _ = u.notify.UserPhone(ctx, id)
		}
	}
	if phone == "" {
		return nil, failNotSent(http.StatusBadRequest, "Nomor WhatsApp tujuan tidak ditemukan.")
	}

	clean := nonDigitRe.ReplaceAllString(phone, "")
	if rest, ok := strings.CutPrefix(clean, "0"); ok {
		clean = "62" + rest
	}

	text, err := u.wa.Send(ctx, provider, gatewayToken, clean, req.Message)
	if err != nil {
		return nil, failNotSent(http.StatusInternalServerError, err.Error())
	}
	return map[string]any{
		"status":    "ok",
		"provider":  provider,
		"recipient": clean,
		"response":  text,
		"sent":      true,
	}, nil
}
