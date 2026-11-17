package usecase

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Status dan Message dikirim apa adanya oleh handler supaya kode dan pesan sama dengan api/*.ts serta PostgREST lama.
type Error struct {
	Status  int
	Message string
	Detail  any
}

func (e *Error) Error() string { return e.Message }

func newError(status int, message string) *Error {
	return &Error{Status: status, Message: message}
}

func BadRequest(message string) *Error { return newError(http.StatusBadRequest, message) }

func Unauthorized(message string) *Error { return newError(http.StatusUnauthorized, message) }

func Forbidden(message string) *Error { return newError(http.StatusForbidden, message) }

func NotFound(message string) *Error { return newError(http.StatusNotFound, message) }

// Pesan sama dengan PostgREST saat WITH CHECK gagal (keputusan 1).
func RLSInsert(table string) *Error {
	return Forbidden(fmt.Sprintf("new row violates row-level security policy for table %q", table))
}

// Teks error.message PostgREST (PGRST116) untuk .single() yang tidak menemukan tepat satu baris.
const msgNotSingle = "Cannot coerce the result to a single JSON object"

// Setara error.message PostgREST: pesan Postgres tanpa awalan "ERROR:" dan SQLSTATE.
func dbMessage(err error) string {
	var pg *pgconn.PgError
	switch {
	case errors.As(err, &pg):
		return pg.Message
	case errors.Is(err, gorm.ErrRecordNotFound):
		return msgNotSingle
	}
	// Error non-Postgres (mis. gagal konek pgx) memuat host dan user DSN, sedangkan supabase-js lama tidak pernah membocorkannya.
	log.Printf("elearning db: %v", err)
	return "Internal server error"
}

// RAISE EXCEPTION dari trigger/RPC (SQLSTATE P0001) dulu sampai ke klien sebagai 400 dengan pesan apa adanya.
func MapDBError(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "P0001" {
		return BadRequest(pg.Message)
	}
	return err
}
