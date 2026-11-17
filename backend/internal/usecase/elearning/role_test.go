package usecase_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestNormalizeTokenRole(t *testing.T) {
	want := map[string][]string{
		"admin":     {"admin", "\tADMIN\n"},
		"kordas":    {"kordas", "koordinator", "korda", "superadmin", "super_admin", "administrator"},
		"asisten":   {"asisten", "assistant", "laboran"},
		"praktikan": {"praktikan", "mahasiswa", "coordinator", "ast", ""},
	}
	for role, inputs := range want {
		for _, in := range inputs {
			assert.Equal(t, role, usecase.NormalizeTokenRole(in), in)
		}
	}
}

func TestNormalizeUIRoleAndCanonical(t *testing.T) {
	cases := map[string][2]string{
		"admin":       {"admin", "koordinator"},
		"superadmin":  {"admin", "koordinator"},
		"super admin": {"admin", "koordinator"},
		"coordinator": {"kordas", "koordinator"},
		"Koordinator": {"kordas", "koordinator"},
		"ast":         {"asisten", "asisten"},
		"laboran":     {"asisten", "asisten"},
		"mahasiswa":   {"praktikan", "mahasiswa"},
		"":            {"praktikan", "mahasiswa"},
	}
	for in, want := range cases {
		assert.Equal(t, want[0], usecase.NormalizeUIRole(in), in)
		assert.Equal(t, want[1], usecase.ToCanonicalRole(in), in)
	}
}

func TestResolveElearningRole(t *testing.T) {
	assert.Equal(t, "admin", usecase.ResolveElearningRole("mahasiswa", " Admin "))
	assert.Equal(t, "admin", usecase.ResolveElearningRole("koordinator", "admin"))
	assert.Equal(t, "kordas", usecase.ResolveElearningRole("koordinator", ""))
	assert.Equal(t, "asisten", usecase.ResolveElearningRole("asisten", "praktikan"))

	assert.True(t, usecase.IsStaff("asisten"))
	assert.True(t, usecase.IsStaff("kordas"))
	assert.False(t, usecase.IsStaff("praktikan"))
	assert.True(t, usecase.IsAdminOrKordas("admin"))
	assert.False(t, usecase.IsAdminOrKordas("asisten"))
}

func TestErrorHelpers(t *testing.T) {
	e := usecase.RLSInsert("student_progress")
	assert.Equal(t, http.StatusForbidden, e.Status)
	assert.Equal(t, `new row violates row-level security policy for table "student_progress"`, e.Message)

	raised := fmt.Errorf("query: %w", &pgconn.PgError{Code: "P0001", Message: "Prestige belum memenuhi syarat"})
	var ue *usecase.Error
	assert.True(t, errors.As(usecase.MapDBError(raised), &ue))
	assert.Equal(t, http.StatusBadRequest, ue.Status)
	assert.Equal(t, "Prestige belum memenuhi syarat", ue.Message)

	other := &pgconn.PgError{Code: "23505"}
	assert.Same(t, other, usecase.MapDBError(other))
}
