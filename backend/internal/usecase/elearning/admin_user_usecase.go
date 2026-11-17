package usecase

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// XP yang dipotong per pelajaran oleh resetLevelProgress/resetLessonProgress, tanpa melihat xp_reward maupun prestige.
const resetXPPerLesson = 60

const (
	msgOwnRole       = "Anda tidak dapat mengubah peran akun Anda sendiri."
	msgManageKordas  = "Hanya Kordas (Super Admin) yang dapat mengelola akun Admin dan Kordas."
	msgNoRoleAuth    = "Anda tidak memiliki otoritas untuk mengubah peran."
	msgAdminArea     = "Maaf, area ini hanya dapat diakses oleh Administrator atau Editor."
	msgResetStaff    = "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres."
	msgResetLevel    = "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres level."
	msgResetLesson   = "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres pelajaran."
	msgAdjustXP      = "Akses Ditolak: Hanya Admin atau Kordas yang dapat menyesuaikan XP."
	msgDeleteAccount = "Akses Ditolak: Hanya Admin atau Kordas yang dapat menghapus pengguna."
)

type AdminUserUsecase struct {
	users        repository.UserRepository
	progress     repository.ProgressRepository
	curriculum   repository.CurriculumRepository
	unlocked     repository.AchievementRepository
	achievements *AchievementUsecase
	logs         *ActivityLogUsecase
	guard        userGuard
}

func NewAdminUserUsecase(users repository.UserRepository, progress repository.ProgressRepository, curriculum repository.CurriculumRepository, unlocked repository.AchievementRepository, achievements *AchievementUsecase, logs *ActivityLogUsecase) *AdminUserUsecase {
	return &AdminUserUsecase{users: users, progress: progress, curriculum: curriculum, unlocked: unlocked, achievements: achievements, logs: logs, guard: userGuard{users: users}}
}

// Peran efektif UI (keputusan 2): resolveElearningRole dari DB, dan bila praktikan memakai peran token, seperti SupabaseProvider.
func (uc *AdminUserUsecase) uiRole(ctx context.Context, p *dto.Principal) (string, error) {
	u, err := uc.users.GetProfile(ctx, p.NIM)
	if err != nil {
		return "", err
	}
	if u != nil {
		if r := ResolveElearningRole(u.Role, strValue(u.ElearningRole)); r != RolePraktikan {
			return r, nil
		}
	}
	return cmp.Or(p.Role, RolePraktikan), nil
}

// Guard service lama dicek sebelum aturan RLS (keputusan 1).
func (uc *AdminUserUsecase) require(ctx context.Context, p *dto.Principal, allowed func(string) bool, message string) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	role, err := uc.uiRole(ctx, p)
	if err != nil {
		return err
	}
	if !allowed(role) {
		return Forbidden(message)
	}
	return nil
}

// AdminDashboard hanya menyembunyikan daftar ini di UI; policy SELECT users membuka semua baris untuk token mana pun.
func (uc *AdminUserUsecase) List(ctx context.Context, p *dto.Principal, page, limit int, q string) (*dto.Page[repository.AdminUser], error) {
	return listPage(p, page, limit, func() ([]repository.AdminUser, int64, error) { return uc.users.ListAdmin(ctx, page, limit, q) })
}

func (uc *AdminUserUsecase) Progress(ctx context.Context, p *dto.Principal, nim string, page, limit int) (*dto.Page[entity.StudentProgress], error) {
	return listPage(p, page, limit, func() ([]entity.StudentProgress, int64, error) {
		if !ownsOrStaff(p, nim) {
			return nil, 0, nil
		}
		return uc.progress.ListByNIM(ctx, nim, page, limit)
	})
}

// Klien mengirim peta penuh hasil siklus auto -> unlocked -> locked, sama dengan handleUpdateUserAccessOverride.
func (uc *AdminUserUsecase) LevelAccess(ctx context.Context, p *dto.Principal, nim string, req *dto.LevelAccessRequest) error {
	if err := uc.require(ctx, p, IsAdminOrKordas, msgAdminArea); err != nil {
		return err
	}
	raw, err := json.Marshal(req.LevelAccessOverrides)
	if err != nil {
		return err
	}
	_, err = uc.guard.update(ctx, p, nim, map[string]any{"level_access_overrides": datatypes.JSON(raw)})
	return err
}

// Disalin dari handleToggleRole (AdminDashboard.tsx:1289). isAdmin dan isCoordinator lama sama-sama admin/kordas.
func (uc *AdminUserUsecase) Role(ctx context.Context, p *dto.Principal, nim string, req *dto.RoleRequest) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if nim == p.NIM {
		return Forbidden(msgOwnRole)
	}
	role, err := uc.uiRole(ctx, p)
	if err != nil {
		return err
	}
	if !IsAdminOrKordas(role) {
		if req.Role == RoleKordas || req.Role == RoleAdmin {
			return Forbidden(msgManageKordas)
		}
		target, err := uc.users.RoleByNIM(ctx, nim)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// Daftar user admin memetakan role mentah dengan `u.role || 'praktikan'` tanpa normalisasi.
		switch cmp.Or(target, RolePraktikan) {
		case RoleKordas, RoleAdmin, CanonicalKoordinator:
			return Forbidden(msgManageKordas)
		}
		return Forbidden(msgNoRoleAuth)
	}
	var elearningRole any
	if req.Role == RoleAdmin {
		elearningRole = RoleAdmin
	}
	_, err = uc.guard.update(ctx, p, nim, map[string]any{"role": ToCanonicalRole(req.Role), "elearning_role": elearningRole})
	return err
}

// Disalin dari resetUserProgress. Delete student_progress dan unlocked_achievements terbuka bagi staf token,
// sedangkan update users baris orang lain hanya admin/kordas, jadi asisten mereset progres tanpa menyentuh XP (T-16).
func (uc *AdminUserUsecase) Reset(ctx context.Context, p *dto.Principal, nim string) error {
	if err := uc.require(ctx, p, IsStaff, msgResetStaff); err != nil {
		return err
	}
	if IsStaff(p.Role) {
		if _, err := uc.progress.DeleteByNIM(ctx, nim); err != nil {
			return MapDBError(err)
		}
		if _, err := uc.unlocked.DeleteByNIM(ctx, nim); err != nil {
			log.Printf("elearning: hapus unlocked_achievements %s: %v", nim, err)
		}
	}
	if _, err := uc.guard.update(ctx, p, nim, map[string]any{"xp": 0, "level": 1, "streak": 0}); err != nil {
		return err
	}
	uc.logs.audit(ctx, p, "progress_reset", "Mereset seluruh progress belajar mahasiswa NIM: "+nim)
	return nil
}

// Disalin dari resetLevelProgress; daftar pelajaran dibaca dari DB, bukan dari kurikulum di store browser.
func (uc *AdminUserUsecase) ResetLevel(ctx context.Context, p *dto.Principal, nim string, req *dto.ResetLevelRequest) error {
	if err := uc.require(ctx, p, IsStaff, msgResetLevel); err != nil {
		return err
	}
	lessonIDs, err := uc.curriculum.LessonIDsByLevel(ctx, req.LevelID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound(fmt.Sprintf("Level %s not found", req.LevelID))
	}
	if err != nil {
		return err
	}
	var done []string
	if ownsOrStaff(p, nim) {
		if done, err = uc.progress.LessonIDsIn(ctx, nim, lessonIDs); err != nil {
			return err
		}
	}
	if len(done) == 0 {
		return nil
	}
	if IsStaff(p.Role) {
		if _, err := uc.progress.DeleteByNIMLessons(ctx, nim, done); err != nil {
			return MapDBError(err)
		}
	}
	return uc.deductXP(ctx, p, nim, len(done), "Mereset progres level \""+req.LevelID+"\" untuk mahasiswa NIM: "+nim)
}

func (uc *AdminUserUsecase) ResetLesson(ctx context.Context, p *dto.Principal, nim string, req *dto.ResetLessonRequest) error {
	if err := uc.require(ctx, p, IsStaff, msgResetLesson); err != nil {
		return err
	}
	if IsStaff(p.Role) {
		if _, err := uc.progress.DeleteByNIMLessons(ctx, nim, []string{req.LessonID}); err != nil {
			return MapDBError(err)
		}
	}
	title := "undefined"
	if req.LessonTitle != nil {
		title = *req.LessonTitle
	}
	return uc.deductXP(ctx, p, nim, 1, "Mereset progres pelajaran \""+title+"\" untuk mahasiswa NIM: "+nim)
}

// .single() yang gagal membuat XP tidak dipotong dan audit tidak ditulis. Error update users tidak diperiksa kode lama.
func (uc *AdminUserUsecase) deductXP(ctx context.Context, p *dto.Principal, nim string, lessons int, details string) error {
	cur, err := uc.users.XPLevel(ctx, nim)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	xp := max(0, cur.XP-lessons*resetXPPerLesson)
	if _, err := uc.guard.update(ctx, p, nim, map[string]any{"xp": xp, "level": calculateLevel(xp)}); err != nil {
		log.Printf("elearning: potong XP %s: %v", nim, err)
	}
	uc.logs.audit(ctx, p, "progress_reset", details)
	return nil
}

// Disalin dari adjustUserXp.
func (uc *AdminUserUsecase) AdjustXP(ctx context.Context, p *dto.Principal, nim string, req *dto.AdjustXPRequest) (*dto.XPResponse, error) {
	if err := uc.require(ctx, p, IsAdminOrKordas, msgAdjustXP); err != nil {
		return nil, err
	}
	xp := max(0, int(math.Floor(*req.XP+0.5)))
	level := calculateLevel(xp)
	oldLevel := 1
	if cur, err := uc.users.XPLevel(ctx, nim); err == nil {
		oldLevel = cmp.Or(cur.Level, 1)
	}
	if _, err := uc.guard.update(ctx, p, nim, map[string]any{"xp": xp, "level": level}); err != nil {
		return nil, err
	}
	uc.logs.audit(ctx, p, "xp_adjusted", fmt.Sprintf("Mengubah XP mahasiswa NIM: %s menjadi %d (Level: %d)", nim, xp, level))
	return &dto.XPResponse{XP: xp, Level: level, LeveledUp: level > oldLevel, Achievement: uc.achievements.checkXP(ctx, p, nim, float64(xp))}, nil
}

// Audit ditulis sebelum menghapus, sama dengan deleteUser.
func (uc *AdminUserUsecase) Delete(ctx context.Context, p *dto.Principal, nim string) error {
	if err := uc.require(ctx, p, IsAdminOrKordas, msgDeleteAccount); err != nil {
		return err
	}
	uc.logs.audit(ctx, p, "user_deleted", "Menghapus akun mahasiswa NIM: "+nim)
	_, err := uc.guard.delete(ctx, p, nim)
	return err
}
