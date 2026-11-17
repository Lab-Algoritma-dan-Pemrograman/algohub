package usecase

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/webpush"
)

type PushSender interface {
	Send(ctx context.Context, sub webpush.Subscription, payload []byte) (int, error)
}

type PushUsecase struct {
	auth   notifyAuth
	push   repository.PushRepository
	sender PushSender
}

func NewPushUsecase(push repository.PushRepository, sender PushSender, tokens *Tokens) *PushUsecase {
	return &PushUsecase{auth: notifyAuth{tokens: tokens}, push: push, sender: sender}
}

func (u *PushUsecase) Send(ctx context.Context, fromCron bool, token string, req dto.SendPushRequest) (map[string]any, *Failure) {
	if f := u.auth.authorize(fromCron, token, "Forbidden: Hanya staf yang boleh mengirim notifikasi"); f != nil {
		return nil, f
	}
	if req.Title == "" || req.Body == "" {
		return nil, fail(http.StatusBadRequest, "Title and body are required")
	}

	subs, err := u.push.List(ctx, pushUserFilter(req.TargetUserID))
	if err != nil {
		return nil, fail(http.StatusInternalServerError, errText(err))
	}
	if len(subs) == 0 {
		return map[string]any{"message": "No target subscriptions found", "sent": 0}, nil
	}

	url := req.URL
	if url == "" {
		url = jadwalJagaURL
	}
	payload, err := encodeJSON(pushPayload{Title: req.Title, Body: req.Body, URL: url})
	if err != nil {
		return nil, fail(http.StatusInternalServerError, err.Error())
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		sent    int
		expired []string
	)
	for _, sub := range subs {
		wg.Go(func() {
			_, err := u.sender.Send(ctx, toWebpush(sub), payload)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				sent++
			case isExpired(err):
				expired = append(expired, sub.Endpoint)
			}
		})
	}
	wg.Wait()

	if len(expired) > 0 {
		_ = u.push.DeleteByEndpoints(ctx, expired)
	}
	return map[string]any{
		"message": "Web Push processing complete",
		"sent":    sent,
		"total":   len(subs),
		"cleaned": len(expired),
	}, nil
}

// Id bulat dikirim sebagai angka; selain itu nilai mentah diteruskan supaya Postgres yang menolak,
// sama dengan .eq("user_id", targetUserId) di kode lama.
func pushUserFilter(target any) any {
	if !truthy(target) {
		return nil
	}
	if id, ok := integerID(target); ok {
		return id
	}
	if s, ok := target.(string); ok {
		return s
	}
	return fmt.Sprint(target)
}

func toWebpush(s entity.PushSubscription) webpush.Subscription {
	return webpush.Subscription{Endpoint: s.Endpoint, P256dh: s.P256dh, Auth: s.Auth}
}
