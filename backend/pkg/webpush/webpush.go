package webpush

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	wp "github.com/SherClockHolmes/webpush-go"
)

// TTL default library web-push Node (4 minggu), supaya push yang tertunda tetap sampai seperti dulu.
const ttlSeconds = 2419200

// ErrNoPrivateKey dikembalikan tanpa request jaringan kalau VAPID private key belum diset.
var ErrNoPrivateKey = errors.New("webpush: VAPID private key belum diset")

// Subscription adalah endpoint dan kunci PushSubscription dari browser.
type Subscription struct {
	Endpoint string
	P256dh   string
	Auth     string
}

// StatusError menandai push service membalas status non-2xx.
type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("webpush: push service membalas %d", e.Code)
}

// Sender mengirim notifikasi Web Push terenkripsi dengan tanda tangan VAPID.
type Sender struct {
	public  string
	private string
	subject string
	client  *http.Client
}

// NewSender membuat Sender; client nil memakai http.Client dengan timeout 50 detik.
func NewSender(public, private, subject string, client *http.Client) *Sender {
	if client == nil {
		// 50s: di bawah maxDuration Vercel Hobby (60s) supaya fungsi sempat membalas.
		client = &http.Client{Timeout: 50 * time.Second}
	}
	return &Sender{public: public, private: private, subject: subject, client: client}
}

// Send mengirim payload ke satu subscription dan mengembalikan status HTTP push service.
func (s *Sender) Send(ctx context.Context, sub Subscription, payload []byte) (int, error) {
	if s.private == "" {
		return 0, ErrNoPrivateKey
	}
	resp, err := wp.SendNotificationWithContext(ctx, payload, &wp.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     wp.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &wp.Options{
		HTTPClient: s.client,
		// Library menambah "mailto:" sendiri untuk subject non-https, jadi prefix dibuang supaya tidak ganda.
		Subscriber:      strings.TrimPrefix(s.subject, "mailto:"),
		TTL:             ttlSeconds,
		VAPIDPublicKey:  s.public,
		VAPIDPrivateKey: s.private,
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, &StatusError{Code: resp.StatusCode}
	}
	return resp.StatusCode, nil
}
