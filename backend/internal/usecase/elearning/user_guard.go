package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

const (
	msgUserGrant       = "permission denied for table users"
	msgRoleChange      = "Akses Ditolak: Anda tidak diperbolehkan mengubah peran (role) Anda sendiri!"
	msgPrestigeColumns = "Akses Ditolak: kolom prestige hanya bisa diubah lewat Prestige resmi (tamat 6/6)."
)

// Column grant UPDATE users untuk role DB authenticated (migration-fix-users-column-grants.sql:9-22, tidak pasti).
var userUpdateGrant = map[string]bool{
	"role": true, "nama": true, "xp": true, "level": true, "streak": true, "study_time": true,
	"last_active": true, "kelas": true, "jurusan": true, "elearning_role": true, "level_access_overrides": true,
}

// Kolom yang dijaga check_user_updates beserta pesan exception-nya (migration-fix-users-column-grants.sql:40).
var guardedUserColumns = []struct{ column, message string }{
	{"level_access_overrides", "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah level_access_overrides!"},
	{"assessment_access", "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah assessment_access!"},
	{"elearning_role", "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah elearning_role!"},
}

// Koneksi Go memakai role postgres sehingga RLS, column grant, dan trigger yang bergantung auth_role() tidak berlaku
// di DB. Semua tulis users atas nama token lewat sini dengan urutan PostgreSQL: grant, USING/WITH CHECK, lalu trigger
// BEFORE secara alfabet (keputusan 1-4, elearning-rls.md § Aturan usecase).
type userGuard struct {
	users repository.UserRepository
}

func (g userGuard) update(ctx context.Context, p *dto.Principal, nim string, cols map[string]any) (int64, error) {
	for k := range cols {
		if !userUpdateGrant[k] {
			return 0, Forbidden(msgUserGrant)
		}
	}
	privileged := IsAdminOrKordas(p.Role)
	if !privileged && nim != p.AuthNIM {
		return 0, nil
	}
	if !privileged && touchesGuardedColumn(cols) {
		old, err := g.users.GetProfile(ctx, nim)
		if err != nil {
			return 0, err
		}
		if old == nil {
			return 0, nil
		}
		if err := checkUserUpdates(p.Role, old, cols); err != nil {
			return 0, err
		}
	}
	rows, err := g.users.UpdateUser(ctx, nim, cols, !privileged)
	return rows, guardDBError(err)
}

func (g userGuard) insert(ctx context.Context, p *dto.Principal, cols map[string]any) error {
	privileged := IsAdminOrKordas(p.Role)
	if nim, _ := cols["nim"].(string); !privileged && nim != p.AuthNIM {
		return RLSInsert("users")
	}
	return MapDBError(g.users.Insert(ctx, cols, !privileged))
}

func (g userGuard) delete(ctx context.Context, p *dto.Principal, nim string) (int64, error) {
	if !IsAdminOrKordas(p.Role) {
		return 0, nil
	}
	rows, err := g.users.Delete(ctx, nim)
	return rows, MapDBError(err)
}

func (g userGuard) incrementStudyTime(ctx context.Context, p *dto.Principal, nim string, seconds int64) (*int64, error) {
	privileged := IsAdminOrKordas(p.Role)
	if !privileged && nim != p.AuthNIM {
		return nil, nil
	}
	total, err := g.users.IncrementStudyTime(ctx, nim, seconds, !privileged)
	return total, guardDBError(err)
}

func guardDBError(err error) error {
	if errors.Is(err, repository.ErrPrestigeGuard) {
		return BadRequest(msgPrestigeColumns)
	}
	return MapDBError(err)
}

func touchesGuardedColumn(cols map[string]any) bool {
	if _, ok := cols["role"]; ok {
		return true
	}
	for _, g := range guardedUserColumns {
		if _, ok := cols[g.column]; ok {
			return true
		}
	}
	return false
}

// Meniru check_user_updates. NEW.role sudah dikanonikkan trigger_canonicalize_user_role yang jalan lebih dulu (alfabet).
func checkUserUpdates(tokenRole string, old *entity.User, cols map[string]any) error {
	if IsAdminOrKordas(tokenRole) {
		return nil
	}
	if v, ok := cols["role"]; ok {
		next, _ := v.(string)
		next = canonicalizeDBRole(next)
		if next != old.Role && next != tokenRole {
			return BadRequest(msgRoleChange)
		}
	}
	oldValues := map[string]any{
		"level_access_overrides": json.RawMessage(old.LevelAccessOverrides),
		"assessment_access":      json.RawMessage(old.AssessmentAccess),
		"elearning_role":         old.ElearningRole,
	}
	for _, g := range guardedUserColumns {
		if v, ok := cols[g.column]; ok && !sameJSON(oldValues[g.column], v) {
			return BadRequest(g.message)
		}
	}
	return nil
}

// Pemetaan canonicalize_user_role (migration-canonical-role.sql:12); nilai lain dibiarkan apa adanya.
func canonicalizeDBRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "kordas", "korda", "koordinator":
		return CanonicalKoordinator
	case "praktikan", "mahasiswa":
		return CanonicalMahasiswa
	case "asisten":
		return CanonicalAsisten
	}
	return role
}

// IS DISTINCT FROM untuk jsonb/text: dibandingkan sebagai nilai JSON, NULL hanya sama dengan NULL.
func sameJSON(a, b any) bool {
	da, errA := decodeJSONValue(a)
	db, errB := decodeJSONValue(b)
	return errA == nil && errB == nil && reflect.DeepEqual(da, db)
}

func decodeJSONValue(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = []byte("null")
	}
	var out any
	err = json.Unmarshal(raw, &out)
	return out, err
}
