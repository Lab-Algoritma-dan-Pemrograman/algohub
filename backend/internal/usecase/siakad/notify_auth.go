package usecase

import (
	"errors"
	"net/http"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/webpush"
)

const jadwalJagaURL = "/jadwal-jaga"

// Pemanggil cron internal (X-Cron-Secret cocok) lolos tanpa token; selain itu wajib token staf.
type notifyAuth struct {
	tokens *Tokens
}

func (a notifyAuth) authorize(fromCron bool, token, forbidden string) *Failure {
	if fromCron {
		return nil
	}
	if token == "" {
		return fail(http.StatusUnauthorized, "Unauthorized: Missing session token")
	}
	claims, err := a.tokens.Verify(token)
	if err != nil {
		return fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	role := ""
	if truthy(claims["role"]) {
		role = claimText(claims["role"])
	}
	if !isNotifyStaff(role) {
		return fail(http.StatusForbidden, forbidden)
	}
	return nil
}

type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// Subscription yang sudah tidak berlaku (404/410) dihapus seperti di kode lama.
func isExpired(err error) bool {
	var se *webpush.StatusError
	return errors.As(err, &se) && (se.Code == http.StatusNotFound || se.Code == http.StatusGone)
}

// Error non-Postgres (mis. gagal konek pgx) memuat user, host, dan database; jangan dikirim ke klien.
func errText(err error) string {
	if msg, ok := dbErrorMessage(err); ok {
		return msg
	}
	return "Internal server error"
}

// Format Date.prototype.toISOString JavaScript.
func isoMillis(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
