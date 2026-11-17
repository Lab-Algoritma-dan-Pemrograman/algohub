package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestCurriculumSaveBuildsRowsLikeSaveFullCurriculum(t *testing.T) {
	f := newLearn(t)
	req := &dto.CurriculumRequest{Levels: []dto.CurriculumLevel{
		{ID: "a", Title: "A", AccessMode: "manual", Locked: true},
		{ID: "b", Title: "B", Modules: []dto.CurriculumModule{
			{ID: "m1", Title: "M1"},
			{ID: "m2", Title: "M2", Lessons: []dto.CurriculumLesson{
				{ID: "x", Quiz: json.RawMessage(`null`), TestCases: json.RawMessage(`[{"input":"1"}]`)},
				{ID: "y", Quiz: json.RawMessage(`{"question":"q","correctAnswer":2}`), Solution: str("s")},
			}},
		}},
	}}
	f.curriculum.On("SaveFull", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		levels := args.Get(1).([]entity.Level)
		modules := args.Get(2).([]entity.Module)
		lessons := args.Get(3).([]entity.Lesson)
		assert.Equal(t, "manual", *levels[0].AccessMode)
		assert.True(t, *levels[0].Locked)
		assert.Equal(t, 1, levels[1].SortOrder)
		assert.Equal(t, []entity.Module{{ID: "m1", LevelID: "b", Title: "M1", SortOrder: 0}, {ID: "m2", LevelID: "b", Title: "M2", SortOrder: 1}}, modules)
		require.Len(t, lessons, 2)
		assert.Equal(t, "m2", lessons[1].ModuleID)
		assert.Equal(t, 1, lessons[1].SortOrder)
		assert.JSONEq(t, `{}`, string(lessons[0].Quiz))
		assert.JSONEq(t, `[{"input":"1"}]`, string(lessons[0].TestCases))
		assert.JSONEq(t, `{"question":"q","correctAnswer":2}`, string(lessons[1].Quiz), "staf menyimpan kunci jawaban apa adanya")
		assert.Equal(t, "s", *lessons[1].Solution)
	})
	require.NoError(t, f.curriculumUC.Save(bg, who(usecase.RoleAdmin), req))
}

func TestCurriculumStudentQuizWithoutAnswerKey(t *testing.T) {
	items := studentLessons(t,
		entity.StudentLesson{ID: "a", Quiz: []byte(`[1,2]`)},
		entity.StudentLesson{ID: "b", Quiz: []byte(`{"question":"q"}`)},
		entity.StudentLesson{ID: "c"})
	assert.JSONEq(t, `[1,2]`, string(items[0].Quiz))
	assert.JSONEq(t, `{"question":"q"}`, string(items[1].Quiz))
	assert.Nil(t, items[2].Quiz)
}
