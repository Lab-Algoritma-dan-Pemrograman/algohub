package usecase

import (
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5/pgconn"

	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

// Body dikirim apa adanya sebagai JSON karena frontend siakad lama membaca field error/details.
type Failure struct {
	Status int
	Body   map[string]any
}

func fail(status int, msg string) *Failure {
	return &Failure{Status: status, Body: map[string]any{"error": msg}}
}

func failDetails(status int, msg, details string) *Failure {
	return &Failure{Status: status, Body: map[string]any{"error": msg, "details": details}}
}

// Pesan Postgres yang cocok disamarkan supaya struktur skema tidak bocor ke klien.
var systemErrorRe = regexp.MustCompile(`(?i)constraint|violates|foreign key|relation|table|syntax|null value|permission denied|does not exist|column|parse`)

// Hanya error yang di kode lama datang dari PostgREST yang punya pesan; sisanya (koneksi) jadi 500.
func dbErrorMessage(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Message, true
	}
	var nf *repository.FunctionNotFoundError
	if errors.As(err, &nf) {
		return nf.Error(), true
	}
	return "", false
}
