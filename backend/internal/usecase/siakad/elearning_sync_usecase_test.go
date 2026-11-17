package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

func decodeFixture[T any](t *testing.T, raw string) []T {
	t.Helper()
	var out []T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("fixture %s: %v", raw, err)
	}
	return out
}

const (
	fixtureLevels = `[
		{"id":"level-3","title":"Level 3","sort_order":3,"access_mode":"unlocked"},
		{"id":"level-1","title":"Level 1","sort_order":1},
		{"id":"level-2","title":"Level 2","sort_order":2,"access_mode":null}
	]`
	fixtureModules = `[{"id":"m1","level_id":"level-1"},{"id":"m2","level_id":"level-2"},{"id":"m3","level_id":"level-3"},{"id":"m4","level_id":null}]`
	fixtureLessons = `[{"id":"a","module_id":"m1"},{"id":"b","module_id":"m1"},{"id":"c","module_id":"m2"},{"id":"d","module_id":"m3"},{"id":"e","module_id":"m4"},{"id":"f","module_id":"tidak-ada"}]`
)

func progressFixture(t *testing.T, done ...string) []entity.ElearningStudentProgress {
	items := make([]string, 0, len(done)+1)
	for _, id := range done {
		items = append(items, `{"lesson_id":"`+id+`","completed":true,"completed_at":null}`)
	}
	items = append(items, `{"lesson_id":"d","completed":false}`)
	return decodeFixture[entity.ElearningStudentProgress](t, "["+strings.Join(items, ",")+"]")
}

func TestComputeProgress(t *testing.T) {
	cases := []struct {
		name      string
		done      []string
		overrides map[string]any
		want      string
	}{
		{"belum mulai: level 2 terkunci, level 3 unlocked", nil, nil,
			`{"completed_lessons":0,"total_lessons":3,"completion_percentage":0,"is_completed":false,"completed_levels":[],"current_level":"Level 1"}`},
		{"level 1 tuntas membuka level 2", []string{"a", "b"}, nil,
			`{"completed_lessons":2,"total_lessons":4,"completion_percentage":50,"is_completed":false,"completed_levels":["Level 1"],"current_level":"Level 2"}`},
		{"override locked", []string{"a", "b", "c"}, map[string]any{"level-2": "locked"},
			`{"completed_lessons":2,"total_lessons":3,"completion_percentage":66.67,"is_completed":false,"completed_levels":["Level 1"],"current_level":"Level 3"}`},
		{"override unlocked melewati level 1", []string{"c"}, map[string]any{"level-2": "unlocked"},
			`{"completed_lessons":1,"total_lessons":4,"completion_percentage":25,"is_completed":false,"completed_levels":["Level 2"],"current_level":"Level 1"}`},
		{"override auto memakai access_mode", []string{"a", "b", "c", "d"}, map[string]any{"level-3": "auto"},
			`{"completed_lessons":4,"total_lessons":4,"completion_percentage":100,"is_completed":true,"completed_levels":["Level 1","Level 2","Level 3"],"current_level":"Belum Mulai"}`},
		{"pembulatan dua desimal", []string{"a"}, nil,
			`{"completed_lessons":1,"total_lessons":3,"completion_percentage":33.33,"is_completed":false,"completed_levels":[],"current_level":"Level 1"}`},
	}
	for _, tc := range cases {
		d := elearningData{
			levels:    decodeFixture[entity.ElearningLevel](t, fixtureLevels),
			modules:   decodeFixture[entity.ElearningModule](t, fixtureModules),
			lessons:   decodeFixture[entity.ElearningLesson](t, fixtureLessons),
			progress:  progressFixture(t, tc.done...),
			overrides: tc.overrides,
		}
		p := computeProgress(&d)
		got := string(mustJSON(t, p))
		want := `{"nim":"","student_name":"",` + strings.TrimPrefix(tc.want, "{")
		want = strings.TrimSuffix(want, "}") + `,"last_accessed_at":null}`
		if got != want {
			t.Errorf("%s:\n got %s\nwant %s", tc.name, got, want)
		}
	}
}

func TestComputeProgressJSSemantics(t *testing.T) {
	d := elearningData{
		levels: decodeFixture[entity.ElearningLevel](t, `[
			{"id":2,"title":"B","sort_order":1},
			{"id":1,"title":"A","sort_order":1},
			{"id":3,"title":"Lama","sort_order":0,"locked":true}
		]`),
		modules:   decodeFixture[entity.ElearningModule](t, `[{"id":10,"level_id":2},{"id":"11","level_id":"1"}]`),
		lessons:   decodeFixture[entity.ElearningLesson](t, `[{"id":5,"module_id":"10"},{"id":6,"module_id":11}]`),
		progress:  decodeFixture[entity.ElearningStudentProgress](t, `[{"lesson_id":"5","completed":true},{"lesson_id":6,"completed":1}]`),
		overrides: map[string]any{},
	}
	p := computeProgress(&d)
	// Level "Lama" terkunci (locked=true) walau sort_order paling awal; urutan B lalu A dipertahankan
	// (sort stabil). Lesson 5 tidak tuntas karena progres menyimpan "5" (string), sehingga A terkunci.
	if p.TotalLessons != 1 || p.CompletedLessons != 0 || p.CurrentLevel != "B" || len(p.CompletedLevels) != 0 {
		t.Fatalf("progress %+v", p)
	}
}

func TestLastAccessed(t *testing.T) {
	progress := decodeFixture[entity.ElearningStudentProgress](t, `[
		{"lesson_id":"a","completed_at":"2026-01-02T03:04:05.678901+07:00"},
		{"lesson_id":"b","completed_at":"2025-12-31T23:59:59+00:00"},
		{"lesson_id":"c","completed_at":null}
	]`)
	cases := []struct {
		name     string
		session  *entity.ElearningSession
		progress []entity.ElearningStudentProgress
		want     string
	}{
		{"heartbeat apa adanya", &entity.ElearningSession{LastHeartbeat: json.RawMessage(`"2026-02-01T10:00:00.5+00:00"`)}, progress, `"2026-02-01T10:00:00.5+00:00"`},
		{"completed_at terbaru", &entity.ElearningSession{LastHeartbeat: json.RawMessage(`null`)}, progress, `"2026-01-01T20:04:05.678Z"`},
		{"tanggal rusak = null", nil, decodeFixture[entity.ElearningStudentProgress](t, `[{"completed_at":"kemarin"},{"completed_at":"2026-01-01T00:00:00Z"}]`), "null"},
		{"semua kosong", nil, decodeFixture[entity.ElearningStudentProgress](t, `[{"completed_at":null}]`), "null"},
		{"tanpa progres", nil, nil, "null"},
	}
	for _, tc := range cases {
		got := lastAccessed(tc.session, tc.progress)
		if got == nil {
			got = json.RawMessage("null")
		}
		if string(got) != tc.want {
			t.Errorf("%s: %s, ingin %s", tc.name, got, tc.want)
		}
	}
}

func newSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func jsonRows(values ...string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"to_json"})
	for _, v := range values {
		rows.AddRow(v)
	}
	return rows
}

func expectElearningReads(mock sqlmock.Sqlmock, nim string) {
	mock.ExpectQuery("SELECT to_json(l) FROM public.levels AS l").
		WillReturnRows(jsonRows(`{"id":"level-1","title":"Level 1","sort_order":1,"locked":false,"extra":[1]}`))
	mock.ExpectQuery("SELECT to_json(p) FROM public.student_progress AS p WHERE p.nim = $1").WithArgs(nim).
		WillReturnRows(jsonRows(`{"lesson_id":"a","completed":true,"completed_at":"2026-01-01T00:00:00+00:00"}`))
	mock.ExpectQuery("SELECT to_json(s) FROM public.active_sessions AS s WHERE s.nim = $1 LIMIT 2").WithArgs(nim).
		WillReturnRows(jsonRows())
	mock.ExpectQuery("SELECT json_build_object('id', id, 'level_id', level_id) FROM public.modules").
		WillReturnRows(jsonRows(`{"id":"m1","level_id":"level-1"}`))
	mock.ExpectQuery("SELECT json_build_object('id', id, 'module_id', module_id) FROM public.lessons").
		WillReturnRows(jsonRows(`{"id":"a","module_id":"m1"}`, `{"id":"b","module_id":"m1"}`))
	mock.ExpectQuery("SELECT to_json(level_access_overrides) FROM public.users WHERE nim = $1 LIMIT 2").WithArgs(nim).
		WillReturnRows(sqlmock.NewRows([]string{"to_json"}).AddRow(nil))
}

const (
	syncRowJSON  = `{"nim":"202711001","student_name":"Budi","completed_lessons":1,"lessons_completed":1,"total_lessons":2,"completion_percentage":50,"is_completed":false,"completed_levels":[],"current_level":"Level 1","last_accessed_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-03-04T05:06:07.089Z"}`
	syncBodyJSON = `{"nim":"202711001","student_name":"Budi","completed_lessons":1,"total_lessons":2,"completion_percentage":50,"is_completed":false,"completed_levels":[],"current_level":"Level 1","last_accessed_at":"2026-01-01T00:00:00.000Z"}`
	progressCols = "nim, student_name, completed_lessons, lessons_completed, total_lessons, completion_percentage, is_completed, completed_levels, current_level, last_accessed_at, updated_at"
	progressSets = "student_name = %s.student_name, completed_lessons = %s.completed_lessons, lessons_completed = %s.lessons_completed, total_lessons = %s.total_lessons, completion_percentage = %s.completion_percentage, is_completed = %s.is_completed, completed_levels = %s.completed_levels, current_level = %s.current_level, last_accessed_at = %s.last_accessed_at, updated_at = %s.updated_at"
	insertSQL    = "INSERT INTO public.elearning_progress (" + progressCols + ") SELECT " + progressCols +
		" FROM json_populate_record(NULL::public.elearning_progress, $1::json)"
)

func sets(alias string) string {
	return strings.ReplaceAll(progressSets, "%s", alias)
}

func newSyncRig(t *testing.T) (uc *ElearningSyncUsecase, elearningMock, siakadMock sqlmock.Sqlmock) {
	t.Helper()
	elearningDB, elearningMock := newSQLMock(t)
	siakadDB, siakadMock := newSQLMock(t)
	uc = NewElearningSyncUsecase(NewTokens(testSecret),
		repository.NewElearningRepository(elearningDB), repository.NewProgressRepository(siakadDB))
	uc.now = func() time.Time { return time.Date(2026, 3, 4, 5, 6, 7, 89e6, time.UTC) }
	return uc, elearningMock, siakadMock
}

func TestSyncElearningReadsElearningWritesSiakad(t *testing.T) {
	uc, elearningMock, siakadMock := newSyncRig(t)
	expectElearningReads(elearningMock, "202711001")
	siakadMock.ExpectExec(insertSQL + " ON CONFLICT (nim) DO UPDATE SET " + sets("EXCLUDED")).
		WithArgs(syncRowJSON).WillReturnResult(sqlmock.NewResult(0, 1))

	token := sign(t, map[string]any{"username": "202711001", "full_name": "Budi", "role": "praktikan"})
	p, f := uc.Sync(context.Background(), token, true)
	if f != nil {
		t.Fatalf("failure %+v", f)
	}
	if got := string(mustJSON(t, p)); got != syncBodyJSON {
		t.Fatalf("body %s", got)
	}
	if err := elearningMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := siakadMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncElearningFallback(t *testing.T) {
	for _, exists := range []bool{true, false} {
		uc, elearningMock, siakadMock := newSyncRig(t)
		expectElearningReads(elearningMock, "202711001")
		siakadMock.ExpectExec(insertSQL + " ON CONFLICT (nim) DO UPDATE SET " + sets("EXCLUDED")).
			WithArgs(syncRowJSON).WillReturnError(errors.New("no unique constraint"))
		idRows := sqlmock.NewRows([]string{"id"})
		if exists {
			idRows.AddRow(1)
		}
		siakadMock.ExpectQuery("SELECT id FROM public.elearning_progress WHERE nim = $1 LIMIT 2").
			WithArgs("202711001").WillReturnRows(idRows)
		if exists {
			siakadMock.ExpectExec("UPDATE public.elearning_progress AS t SET " + sets("r") +
				" FROM json_populate_record(NULL::public.elearning_progress, $1::json) AS r WHERE t.nim = r.nim").
				WithArgs(syncRowJSON).WillReturnError(errors.New("diabaikan"))
		} else {
			siakadMock.ExpectExec(insertSQL).WithArgs(syncRowJSON).WillReturnError(errors.New("diabaikan"))
		}

		token := sign(t, map[string]any{"nim": "202711001", "nama": "Budi"})
		if _, f := uc.Sync(context.Background(), token, true); f != nil {
			t.Fatalf("exists=%v: failure %+v", exists, f)
		}
		if err := siakadMock.ExpectationsWereMet(); err != nil {
			t.Fatalf("exists=%v: %v", exists, err)
		}
	}
}

type fakeElearning struct {
	failAt string
}

func (f fakeElearning) err(step string) error {
	if f.failAt == step {
		return errors.New("db error")
	}
	return nil
}

func (f fakeElearning) Levels(context.Context) ([]entity.ElearningLevel, error) {
	return nil, f.err("levels")
}

func (f fakeElearning) StudentProgress(context.Context, string) ([]entity.ElearningStudentProgress, error) {
	return nil, f.err("progress")
}

func (f fakeElearning) ActiveSession(context.Context, string) (*entity.ElearningSession, error) {
	return nil, errors.New("diabaikan")
}

func (f fakeElearning) Modules(context.Context) ([]entity.ElearningModule, error) {
	return nil, f.err("modules")
}

func (f fakeElearning) Lessons(context.Context) ([]entity.ElearningLesson, error) {
	return nil, f.err("lessons")
}

func (f fakeElearning) LevelOverrides(context.Context, string) (json.RawMessage, error) {
	return json.RawMessage(`[1]`), errors.New("diabaikan")
}

type fakeProgress struct {
	upserts int
}

func (f *fakeProgress) Upsert(context.Context, *entity.ElearningProgressRow) error {
	f.upserts++
	return nil
}

func (f *fakeProgress) ExistsByNIM(context.Context, string) (bool, error) { return false, nil }

func (f *fakeProgress) Update(context.Context, *entity.ElearningProgressRow) error { return nil }

func (f *fakeProgress) Insert(context.Context, *entity.ElearningProgressRow) error { return nil }

func TestSyncElearningFailures(t *testing.T) {
	nimToken := sign(t, map[string]any{"nim": "1"})
	cases := []struct {
		name      string
		token     string
		hasBearer bool
		failAt    string
		status    int
		msg       string
	}{
		{"tanpa bearer", "", false, "", 401, "Unauthorized: Missing session token"},
		{"token rusak", "x.y.z", true, "", 401, "Unauthorized: Invalid or expired token"},
		{"tanpa nim", sign(t, map[string]any{"role": "praktikan"}), true, "", 400, "NIM pengguna tidak ditemukan di token sesi"},
		{"levels", nimToken, true, "levels", 502, "Gagal mengambil data kurikulum dari E-Learning"},
		{"progress", nimToken, true, "progress", 502, "Gagal mengambil progres belajar dari E-Learning"},
		{"modules", nimToken, true, "modules", 502, "Gagal mengambil struktur modul dari E-Learning"},
		{"lessons", nimToken, true, "lessons", 502, "Gagal mengambil daftar pelajaran dari E-Learning"},
	}
	for _, tc := range cases {
		progress := &fakeProgress{}
		uc := NewElearningSyncUsecase(NewTokens(testSecret), fakeElearning{failAt: tc.failAt}, progress)
		_, f := uc.Sync(context.Background(), tc.token, tc.hasBearer)
		if f == nil || f.Status != tc.status || f.Body["error"] != tc.msg || progress.upserts != 0 {
			t.Errorf("%s: %+v, upsert %d", tc.name, f, progress.upserts)
		}
	}

	progress := &fakeProgress{}
	uc := NewElearningSyncUsecase(NewTokens(testSecret), fakeElearning{}, progress)
	p, f := uc.Sync(context.Background(), sign(t, map[string]any{"nim": json.Number("42")}), true)
	if f != nil || p.NIM != "42" || p.StudentName != "Mahasiswa" || p.CurrentLevel != "Belum Mulai" || progress.upserts != 1 {
		t.Fatalf("default: %+v %+v upsert %d", p, f, progress.upserts)
	}
}
