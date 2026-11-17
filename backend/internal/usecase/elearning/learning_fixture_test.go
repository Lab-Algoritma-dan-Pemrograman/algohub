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
	games        *mocks.GameRepository
	board        *mocks.LeaderboardRepository
	prestige     *mocks.PrestigeRepository
	playground   *mocks.PlaygroundRepository
	sessions     *mocks.SessionRepository
	uploader     *fakeUploader

	curriculumUC  *usecase.CurriculumUsecase
	progressUC    *usecase.ProgressUsecase
	profileUC     *usecase.ProfileUsecase
	achievementUC *usecase.AchievementUsecase
	adminUC       *usecase.AdminUserUsecase
	logUC         *usecase.ActivityLogUsecase
	gameUC        *usecase.GameUsecase
	boardUC       *usecase.LeaderboardUsecase
	prestigeUC    *usecase.PrestigeUsecase
	playgroundUC  *usecase.PlaygroundUsecase
	sessionUC     *usecase.SessionUsecase
	storageUC     *usecase.StorageUsecase
}

// Urutan soal dibalik supaya hasil acak bisa diperiksa.
func reverseShuffle(n int, swap func(i, j int)) {
	for i := 0; i < n/2; i++ {
		swap(i, n-1-i)
	}
}

type fakeUploader struct {
	path, contentType string
	upsert, called    bool
	err               error
}

func (u *fakeUploader) Upload(path string, _ []byte, contentType string, upsert bool) (string, error) {
	u.path, u.contentType, u.upsert, u.called = path, contentType, upsert, true
	if u.err != nil {
		return "", u.err
	}
	return "https://storage.test/" + path, nil
}

func newLearn(t *testing.T) *learnFixture {
	t.Helper()
	f := &learnFixture{
		users:        mocks.NewUserRepository(t),
		curriculum:   mocks.NewCurriculumRepository(t),
		progress:     mocks.NewProgressRepository(t),
		achievements: mocks.NewAchievementRepository(t),
		logs:         mocks.NewActivityLogRepository(t),
		games:        mocks.NewGameRepository(t),
		board:        mocks.NewLeaderboardRepository(t),
		prestige:     mocks.NewPrestigeRepository(t),
		playground:   mocks.NewPlaygroundRepository(t),
		sessions:     mocks.NewSessionRepository(t),
		uploader:     &fakeUploader{},
	}
	now := func() time.Time { return learnNow }
	f.curriculumUC = usecase.NewCurriculumUsecase(f.curriculum)
	f.logUC = usecase.NewActivityLogUsecase(f.logs, now)
	f.achievementUC = usecase.NewAchievementUsecase(f.achievements, f.users, f.progress, now)
	f.progressUC = usecase.NewProgressUsecase(f.users, f.progress, f.curriculum, f.achievementUC, now)
	f.profileUC = usecase.NewProfileUsecase(f.users, f.achievementUC, now)
	f.adminUC = usecase.NewAdminUserUsecase(f.users, f.progress, f.curriculum, f.achievements, f.achievementUC, f.logUC)
	f.gameUC = usecase.NewGameUsecase(f.games, f.users, now, reverseShuffle)
	f.boardUC = usecase.NewLeaderboardUsecase(f.board, now)
	f.prestigeUC = usecase.NewPrestigeUsecase(f.users, f.curriculum, f.progress, f.prestige, f.logUC, now)
	f.playgroundUC = usecase.NewPlaygroundUsecase(f.playground)
	f.sessionUC = usecase.NewSessionUsecase(f.sessions, f.users, f.progress, f.progressUC, now)
	f.storageUC = usecase.NewStorageUsecase(f.uploader, now)
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
