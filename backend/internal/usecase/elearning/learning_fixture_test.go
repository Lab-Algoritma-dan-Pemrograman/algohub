package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// 17 November 2026 10:00 WIB.
var learnNow = time.Date(2026, 11, 17, 3, 0, 0, 0, time.UTC)

type learnFixture struct {
	users        *mocks.UserRepository
	curriculum   *mocks.CurriculumRepository
	progress     *mocks.ProgressRepository
	achievements *mocks.AchievementRepository
	logs         *mocks.ActivityLogRepository

	curriculumUC  *usecase.CurriculumUsecase
	progressUC    *usecase.ProgressUsecase
	profileUC     *usecase.ProfileUsecase
	achievementUC *usecase.AchievementUsecase
	adminUC       *usecase.AdminUserUsecase
	logUC         *usecase.ActivityLogUsecase
}

func newLearn(t *testing.T) *learnFixture {
	t.Helper()
	f := &learnFixture{
		users:        mocks.NewUserRepository(t),
		curriculum:   mocks.NewCurriculumRepository(t),
		progress:     mocks.NewProgressRepository(t),
		achievements: mocks.NewAchievementRepository(t),
		logs:         mocks.NewActivityLogRepository(t),
	}
	now := func() time.Time { return learnNow }
	f.curriculumUC = usecase.NewCurriculumUsecase(f.curriculum)
	f.logUC = usecase.NewActivityLogUsecase(f.logs, now)
	f.achievementUC = usecase.NewAchievementUsecase(f.achievements, f.users, f.progress, now)
	f.progressUC = usecase.NewProgressUsecase(f.users, f.progress, f.curriculum, f.achievementUC, now)
	f.profileUC = usecase.NewProfileUsecase(f.users, f.achievementUC, now)
	f.adminUC = usecase.NewAdminUserUsecase(f.users, f.progress, f.curriculum, f.achievements, f.achievementUC, f.logUC)
	return f
}

func who(role string) *dto.Principal {
	return &dto.Principal{NIM: nimSelf, AuthNIM: nimSelf, Nama: "Nama " + nimSelf, Kelas: "TI-1", Role: role}
}

// Principal yang nim tampilannya berbeda dari auth_nim(), misal token tanpa claim nim dengan username != sub.
func mismatched(role string) *dto.Principal {
	p := who(role)
	p.AuthNIM = "sub-lain"
	return p
}

func str(s string) *string { return &s }

// Profil di DB yang menentukan peran efektif UI (keputusan 2).
func (f *learnFixture) profileRole(nim, role string, elearningRole *string) {
	f.users.On("GetProfile", mock.Anything, nim).Return(&entity.User{NIM: nim, Nama: "Nama " + nim, Role: role, ElearningRole: elearningRole, PrestigeMult: 1}, nil)
}

func (f *learnFixture) expectAudit(eventType, details string) {
	f.logs.On("Insert", mock.Anything, mock.MatchedBy(func(l *entity.ActivityLog) bool {
		return *l.NIM == nimSelf && l.EventType == eventType && *l.Details == details && l.Timestamp.Equal(learnNow)
	})).Return(nil).Once()
}
