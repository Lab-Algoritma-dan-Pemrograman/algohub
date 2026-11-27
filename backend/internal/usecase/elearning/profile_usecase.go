package usecase

import (
	"cmp"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"gorm.io/datatypes"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

var defaultAssessmentAccess = datatypes.JSON(`{"pre_test":false,"post_test":false,"program_keterampilan":false,"ujian_praktik":false}`)

type ProfileUsecase struct {
	users        repository.UserRepository
	achievements *AchievementUsecase
	guard        userGuard
	now          func() time.Time
}

func NewProfileUsecase(users repository.UserRepository, achievements *AchievementUsecase, now func() time.Time) *ProfileUsecase {
	return &ProfileUsecase{users: users, achievements: achievements, guard: userGuard{users: users}, now: now}
}

// Disalin dari initialize di SupabaseProvider.tsx:84-260. Error baca, insert, dan update ditelan seperti kode lama
// (contoh nyata: update yang ikut menulis shift ditolak column grant, T-26), jadi sinkron tidak pernah gagal.
func (uc *ProfileUsecase) Sync(ctx context.Context, p *dto.Principal) (*dto.ProfileSyncResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	now := uc.now().UTC()
	u, err := uc.users.GetProfile(ctx, p.NIM)
	if err != nil {
		log.Printf("elearning: baca profil: %v", err)
	}
	if err != nil || u == nil {
		return &dto.ProfileSyncResponse{Profile: uc.create(ctx, p, now), NewlyUnlocked: []dto.Achievement{}}, nil
	}
	return uc.refresh(ctx, p, u, now), nil
}

func (uc *ProfileUsecase) create(ctx context.Context, p *dto.Principal, now time.Time) dto.Profile {
	role := cmp.Or(p.Role, RolePraktikan)
	jurusan, email := nonEmpty(p.Jurusan), nonEmpty(p.Email)
	cols := map[string]any{
		"nim": p.NIM, "nama": p.Nama, "kelas": p.Kelas, "jurusan": jurusan, "shift": p.Shift, "email": email,
		"xp": 0, "level": 1, "streak": 1, "study_time": 0, "last_active": now, "created_at": now, "role": role,
		"assessment_access": defaultAssessmentAccess, "level_access_overrides": datatypes.JSON(`{}`),
	}
	if err := uc.guard.insert(ctx, p, cols); err != nil {
		log.Printf("elearning: insert profil %s: %v", p.NIM, err)
	}
	return dto.Profile{
		NIM: p.NIM, Nama: p.Nama, Kelas: p.Kelas, Jurusan: jurusan, Shift: p.Shift, Email: email,
		PrestigeMult: 1, XP: 0, Level: 1, Streak: 1, LastActive: &now, CreatedAt: &now, Role: role,
		AssessmentAccess: json.RawMessage(defaultAssessmentAccess), LevelAccessOverrides: json.RawMessage(`{}`),
	}
}

func (uc *ProfileUsecase) refresh(ctx context.Context, p *dto.Principal, u *entity.User, now time.Time) *dto.ProfileSyncResponse {
	streak, _ := calculateStreak(u.LastActive, u.Streak, now)
	resolved := ResolveElearningRole(u.Role, strValue(u.ElearningRole))
	effective := resolved
	if resolved == RolePraktikan {
		effective = cmp.Or(p.Role, RolePraktikan)
	}
	overrides := json.RawMessage(u.LevelAccessOverrides)
	if !jsTruthy(parsedOrNil(overrides)) {
		overrides = json.RawMessage(`{}`)
	}
	prof := dto.Profile{
		NIM:                  u.NIM,
		Nama:                 cmp.Or(p.Nama, u.Nama),
		Kelas:                cmp.Or(p.Kelas, u.Kelas),
		Jurusan:              firstNonEmpty(u.Jurusan, p.Jurusan),
		Shift:                firstNonEmpty(u.Shift, p.Shift),
		Email:                firstNonEmpty(u.Email, p.Email),
		PrestigeTier:         nonEmpty(u.PrestigeTier),
		PrestigeMult:         cmp.Or(u.PrestigeMult, 1),
		PrestigeCount:        u.PrestigeCount,
		XP:                   u.XP,
		Level:                u.Level,
		Streak:               streak,
		LastActive:           u.LastActive,
		CreatedAt:            u.CreatedAt,
		Role:                 effective,
		AssessmentAccess:     json.RawMessage(u.AssessmentAccess),
		LevelAccessOverrides: overrides,
	}
	if u.StudyTime != nil {
		prof.StudyTime = *u.StudyTime
	}

	updates := map[string]any{"last_active": now, "streak": streak}
	// Role token tidak ditulis balik bila elearning_role memberi peran lebih tinggi, karena token membawa peran kanonik.
	hasRoleChange := p.Role != "" && resolved == RolePraktikan && NormalizeUIRole(u.Role) != p.Role
	hasJurusanChange := nonEmpty(p.Jurusan) != nil && !sameString(prof.Jurusan, p.Jurusan)
	hasShiftChange := p.Shift != nil && !sameString(prof.Shift, p.Shift)
	if prof.Nama != p.Nama {
		updates["nama"] = p.Nama
	}
	if kelas := strings.TrimSpace(p.Kelas); kelas != "" && prof.Kelas != kelas {
		updates["kelas"] = kelas // T-30: kelas kosong mengosongkan kelas_id/class_code lewat trigger.
	}
	if hasRoleChange {
		updates["role"] = p.Role
	}
	if hasJurusanChange {
		updates["jurusan"] = *p.Jurusan
	}
	if hasShiftChange {
		updates["shift"] = *p.Shift
	}
	if _, err := uc.guard.update(ctx, p, p.NIM, updates); err != nil {
		log.Printf("elearning: sinkron profil %s: %v", p.NIM, err)
	}

	prof.LastActive = &now
	prof.Nama = p.Nama
	prof.Kelas = p.Kelas
	if nonEmpty(p.Jurusan) != nil {
		prof.Jurusan = p.Jurusan
	}
	if hasShiftChange {
		prof.Shift = p.Shift
	}

	newly := []dto.Achievement{}
	if streak >= 3 {
		newly = uc.achievements.unlock(ctx, p, achievementUser{nim: prof.NIM, xp: float64(prof.XP), streak: streak}, &dto.AchievementCheckRequest{})
	}
	return &dto.ProfileSyncResponse{Profile: prof, NewlyUnlocked: newly}
}

// Operator `x || null` untuk string opsional.
func nonEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func firstNonEmpty(a, b *string) *string {
	if v := nonEmpty(a); v != nil {
		return v
	}
	return nonEmpty(b)
}

func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parsedOrNil(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	v, err := parseJS(raw)
	if err != nil {
		return nil
	}
	return v
}
