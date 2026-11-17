package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestHeartbeatDefaultsAndStudyTimeBatch(t *testing.T) {
	f := newLearn(t)
	p := &dto.Principal{NIM: nimSelf, AuthNIM: nimSelf, Role: usecase.RoleAdmin}
	s := &entity.ActiveSession{NIM: nimSelf, Nama: "Anonymous", Kelas: "Unknown", LastHeartbeat: &learnNow, CurrentActivity: str("lesson")}
	f.sessions.On("Upsert", mock.Anything, s, false).Return(nil)
	total := int64(930)
	f.users.On("IncrementStudyTime", mock.Anything, nimSelf, int64(30), false).Return(&total, nil)
	res, err := f.sessionUC.Heartbeat(bg, p, &dto.HeartbeatRequest{CurrentActivity: str("lesson"), StudySeconds: func() *int64 { v := int64(30); return &v }()})
	require.NoError(t, err)
	assert.Equal(t, &dto.HeartbeatResponse{SessionSaved: true, StudyTime: &dto.StudyTimeResponse{StudyTime: &total}}, res)
}

func TestHeartbeatSwallowsSessionError(t *testing.T) {
	f := newLearn(t)
	f.sessions.On("Upsert", mock.Anything, mock.Anything, false).Return(errors.New("db mati"))
	zero := int64(0)
	res, err := f.sessionUC.Heartbeat(bg, who(usecase.RoleAsisten), &dto.HeartbeatRequest{StudySeconds: &zero})
	require.NoError(t, err)
	assert.Equal(t, &dto.HeartbeatResponse{}, res)
}

func TestMonitoringProgressScope(t *testing.T) {
	f := newLearn(t)
	rows := []repository.NIMLessons{{NIM: nimSelf, LessonIDs: []string{"l1"}}}
	f.progress.On("GroupedCompleted", mock.Anything, nimSelf, 1, 100).Return(rows, int64(1), nil)
	f.progress.On("GroupedCompleted", mock.Anything, "", 2, 50).Return(rows, int64(51), nil)
	page, err := f.sessionUC.Progress(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, rows, page.Items)
	page, err = f.sessionUC.Progress(bg, who(usecase.RoleAsisten), 2, 50)
	require.NoError(t, err)
	assert.Equal(t, int64(51), page.Total)
}

func TestMonitoringStudents(t *testing.T) {
	f := newLearn(t)
	f.users.On("ListStudents", mock.Anything, 1, 100).Return([]repository.MonitoringStudent{{NIM: nimOther}}, int64(1), nil)
	page, err := f.sessionUC.Students(bg, who(usecase.RoleAsisten), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, nimOther, page.Items[0].NIM)
	_, err = f.sessionUC.Students(bg, nil, 1, 100)
	assertMissingToken(t, err)
}
