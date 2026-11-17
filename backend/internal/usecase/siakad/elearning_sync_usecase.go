package usecase

import (
	"cmp"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"slices"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

const (
	accessLocked   = "locked"
	accessUnlocked = "unlocked"
	accessAuto     = "auto"
)

type ElearningSyncUsecase struct {
	tokens    *Tokens
	elearning repository.ElearningRepository
	progress  repository.ProgressRepository
	now       func() time.Time
}

func NewElearningSyncUsecase(tokens *Tokens, elearning repository.ElearningRepository, progress repository.ProgressRepository) *ElearningSyncUsecase {
	return &ElearningSyncUsecase{tokens: tokens, elearning: elearning, progress: progress, now: time.Now}
}

type elearningData struct {
	levels    []entity.ElearningLevel
	progress  []entity.ElearningStudentProgress
	session   *entity.ElearningSession
	modules   []entity.ElearningModule
	lessons   []entity.ElearningLesson
	overrides map[string]any
}

// T-15: hanya header Bearer, sama dengan kode lama.
func (u *ElearningSyncUsecase) Sync(ctx context.Context, token string, hasBearer bool) (*entity.ElearningProgress, *Failure) {
	if !hasBearer {
		return nil, fail(http.StatusUnauthorized, "Unauthorized: Missing session token")
	}
	claims, err := u.tokens.Verify(token)
	if err != nil {
		return nil, fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	nim := firstClaim(claims, "", "nim", "username")
	if nim == "" {
		return nil, fail(http.StatusBadRequest, "NIM pengguna tidak ditemukan di token sesi")
	}
	name := firstClaim(claims, "Mahasiswa", "nama", "full_name")

	data, f := u.load(ctx, nim)
	if f != nil {
		return nil, f
	}
	p := computeProgress(&data)
	p.NIM, p.StudentName = nim, name

	row := &entity.ElearningProgressRow{
		NIM: p.NIM, StudentName: p.StudentName, CompletedLessons: p.CompletedLessons, LessonsCompleted: p.CompletedLessons,
		TotalLessons: p.TotalLessons, CompletionPercentage: p.CompletionPercentage, IsCompleted: p.IsCompleted,
		CompletedLevels: p.CompletedLevels, CurrentLevel: p.CurrentLevel, LastAccessedAt: p.LastAccessedAt,
		UpdatedAt: isoMillis(u.now()),
	}
	// T-18: error upsert dan fallback-nya diabaikan seperti kode lama; respons tetap 200.
	if u.progress.Upsert(ctx, row) != nil {
		if exists, _ := u.progress.ExistsByNIM(ctx, nim); exists {
			_ = u.progress.Update(ctx, row)
		} else {
			_ = u.progress.Insert(ctx, row)
		}
	}
	return p, nil
}

func (u *ElearningSyncUsecase) load(ctx context.Context, nim string) (elearningData, *Failure) {
	var (
		d   elearningData
		err error
	)
	if d.levels, err = u.elearning.Levels(ctx); err != nil {
		return d, fail(http.StatusBadGateway, "Gagal mengambil data kurikulum dari E-Learning")
	}
	if d.progress, err = u.elearning.StudentProgress(ctx, nim); err != nil {
		return d, fail(http.StatusBadGateway, "Gagal mengambil progres belajar dari E-Learning")
	}
	d.session, _ = u.elearning.ActiveSession(ctx, nim)
	if d.modules, err = u.elearning.Modules(ctx); err != nil {
		return d, fail(http.StatusBadGateway, "Gagal mengambil struktur modul dari E-Learning")
	}
	if d.lessons, err = u.elearning.Lessons(ctx); err != nil {
		return d, fail(http.StatusBadGateway, "Gagal mengambil daftar pelajaran dari E-Learning")
	}
	raw, _ := u.elearning.LevelOverrides(ctx, nim)
	if json.Unmarshal(raw, &d.overrides) != nil || d.overrides == nil {
		d.overrides = map[string]any{}
	}
	return d, nil
}

// Port `a || b || fallback` atas klaim token.
func firstClaim(claims map[string]any, fallback string, keys ...string) string {
	for _, k := range keys {
		if v := claims[k]; truthy(v) {
			return claimText(v)
		}
	}
	return fallback
}

// Kunci objek JavaScript selalu string, jadi id 1 dan "1" menunjuk entri yang sama.
func jsKey(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	if len(raw) == 0 {
		return "undefined"
	}
	return string(raw)
}

func rawTruthy(raw json.RawMessage) bool {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return false
	}
	return truthy(v)
}

// Basis persentase sama dengan beranda E-Learning: hanya level yang tidak terkunci yang dihitung.
func computeProgress(d *elearningData) *entity.ElearningProgress {
	moduleToLevel := make(map[string]json.RawMessage, len(d.modules))
	for _, m := range d.modules {
		moduleToLevel[jsKey(m.ID)] = m.LevelID
	}
	lessonsByLevel := map[string][]string{}
	for _, l := range d.lessons {
		lvl, ok := moduleToLevel[jsKey(l.ModuleID)]
		if !ok || !rawTruthy(lvl) {
			continue
		}
		k := jsKey(lvl)
		lessonsByLevel[k] = append(lessonsByLevel[k], string(l.ID))
	}
	// Set JavaScript membandingkan nilai tanpa konversi tipe, jadi id dibandingkan sebagai JSON mentah.
	completed := map[string]bool{}
	for _, p := range d.progress {
		if truthy(p.Completed) {
			completed[string(p.LessonID)] = true
		}
	}

	levels := slices.Clone(d.levels)
	slices.SortStableFunc(levels, func(a, b entity.ElearningLevel) int {
		return cmp.Compare(sortOrder(a), sortOrder(b))
	})
	lock := levelLock{overrides: d.overrides, lessonsByLevel: lessonsByLevel, completed: completed, levels: levels}

	var (
		open, openDone int
		current        *string
	)
	completedLevels := []*string{}
	for i, lvl := range levels {
		if lock.locked(i) {
			continue
		}
		ids := lessonsByLevel[jsKey(lvl.ID)]
		done := countDone(ids, completed)
		open += len(ids)
		openDone += done
		switch {
		case len(ids) > 0 && done >= len(ids):
			completedLevels = append(completedLevels, lvl.Title)
		case done > 0 && !titleSet(current):
			current = lvl.Title
		}
	}
	if !titleSet(current) {
		for i, lvl := range levels {
			ids := lessonsByLevel[jsKey(lvl.ID)]
			if !lock.locked(i) && len(ids) > 0 && countDone(ids, completed) < len(ids) {
				current = lvl.Title
				break
			}
		}
	}

	pct := 0.0
	if open > 0 {
		pct = math.Min(100, math.Floor(float64(openDone)/float64(open)*10000+0.5)/100)
	}
	currentLevel := "Belum Mulai"
	if titleSet(current) {
		currentLevel = *current
	}
	return &entity.ElearningProgress{
		CompletedLessons:     openDone,
		TotalLessons:         open,
		CompletionPercentage: pct,
		IsCompleted:          open > 0 && openDone >= open,
		CompletedLevels:      completedLevels,
		CurrentLevel:         currentLevel,
		LastAccessedAt:       lastAccessed(d.session, d.progress),
	}
}

type levelLock struct {
	overrides      map[string]any
	lessonsByLevel map[string][]string
	completed      map[string]bool
	levels         []entity.ElearningLevel
}

func (l levelLock) effectiveMode(lvl entity.ElearningLevel) any {
	if ov := l.overrides[jsKey(lvl.ID)]; truthy(ov) && ov != accessAuto {
		return ov
	}
	if truthy(lvl.AccessMode) {
		return lvl.AccessMode
	}
	if lvl.Locked == true {
		return accessLocked
	}
	return accessAuto
}

// Mode auto: semua level sebelumnya harus tuntas, kecuali level yang dibuka paksa.
func (l levelLock) locked(idx int) bool {
	lvl := l.levels[idx]
	switch ov := l.overrides[jsKey(lvl.ID)]; ov {
	case accessLocked:
		return true
	case accessUnlocked:
		return false
	}
	switch lvl.AccessMode {
	case accessLocked:
		return true
	case accessUnlocked:
		return false
	}
	if lvl.Locked == true {
		return true
	}
	for _, prev := range l.levels[:idx] {
		if l.effectiveMode(prev) == accessUnlocked {
			continue
		}
		ids := l.lessonsByLevel[jsKey(prev.ID)]
		if countDone(ids, l.completed) < len(ids) {
			return true
		}
	}
	return false
}

func countDone(ids []string, completed map[string]bool) int {
	n := 0
	for _, id := range ids {
		if completed[id] {
			n++
		}
	}
	return n
}

func titleSet(t *string) bool {
	return t != nil && *t != ""
}

func sortOrder(l entity.ElearningLevel) float64 {
	if l.SortOrder == nil {
		return 0
	}
	return *l.SortOrder
}

// last_heartbeat dikirim apa adanya; kalau kosong, completed_at terbaru dalam format toISOString.
// Satu tanggal yang tidak bisa di-parse membuat Math.max bernilai NaN, sehingga hasilnya null.
func lastAccessed(session *entity.ElearningSession, progress []entity.ElearningStudentProgress) json.RawMessage {
	if session != nil && rawTruthy(session.LastHeartbeat) {
		return session.LastHeartbeat
	}
	if len(progress) == 0 {
		return nil
	}
	maxMs := math.Inf(-1)
	for _, p := range progress {
		maxMs = math.Max(maxMs, jsDateMillis(p.CompletedAt))
	}
	if !(maxMs > 0) {
		return nil
	}
	ts, _ := json.Marshal(isoMillis(time.UnixMilli(int64(maxMs))))
	return ts
}

func jsDateMillis(v any) float64 {
	if !truthy(v) {
		return 0
	}
	switch x := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, x)
		if err != nil {
			return math.NaN()
		}
		return float64(t.UnixMilli())
	case float64:
		return x
	}
	return math.NaN()
}
