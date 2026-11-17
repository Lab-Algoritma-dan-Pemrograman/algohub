package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

type fakeSSO struct {
	userID       *int64
	found        bool
	handshakeErr error
	user         *entity.SSOUser
	userErr      error
	student      entity.ScheduleClass
	assistant    entity.ScheduleClass
	gotNow       time.Time
	deleted      []string
	lookups      []string
}

func (f *fakeSSO) HandshakeUserID(_ context.Context, _ string, now time.Time) (userID *int64, found bool, err error) {
	f.gotNow = now
	return f.userID, f.found, f.handshakeErr
}

func (f *fakeSSO) User(context.Context, int64) (entity.SSOUser, bool, error) {
	if f.user == nil {
		return entity.SSOUser{}, false, f.userErr
	}
	return *f.user, true, f.userErr
}

func (f *fakeSSO) DeleteHandshake(_ context.Context, code string) error {
	f.deleted = append(f.deleted, code)
	return errors.New("diabaikan")
}

func (f *fakeSSO) StudentSchedule(context.Context, int64) (entity.ScheduleClass, error) {
	f.lookups = append(f.lookups, "student")
	return f.student, nil
}

func (f *fakeSSO) AssistantSchedule(context.Context, int64) (entity.ScheduleClass, error) {
	f.lookups = append(f.lookups, "assistant")
	return f.assistant, errors.New("diabaikan")
}

func ptr[T any](v T) *T { return &v }

func TestGenerateJWTFailures(t *testing.T) {
	user := &entity.SSOUser{NIM: "202711001", FullName: "Budi"}
	cases := []struct {
		name   string
		code   string
		repo   *fakeSSO
		status int
		msg    string
	}{
		{"tanpa kode", "", &fakeSSO{}, 401, "Missing handshake code"},
		{"kadaluarsa", "abc", &fakeSSO{}, 401, "Sesi verifikasi kadaluarsa. Harap coba lagi."},
		{"error handshake", "abc", &fakeSSO{userID: ptr(int64(1)), found: true, handshakeErr: errors.New("x")}, 401, "Sesi verifikasi kadaluarsa. Harap coba lagi."},
		{"user_id null", "abc", &fakeSSO{found: true, user: user}, 404, "Data pengguna tidak ditemukan di database"},
		{"user tidak ada", "abc", &fakeSSO{userID: ptr(int64(1)), found: true}, 404, "Data pengguna tidak ditemukan di database"},
		{"error user", "abc", &fakeSSO{userID: ptr(int64(1)), found: true, user: user, userErr: errors.New("x")}, 404, "Data pengguna tidak ditemukan di database"},
		{"nama kosong", "abc", &fakeSSO{userID: ptr(int64(1)), found: true, user: &entity.SSOUser{NIM: "1"}}, 400, "Data pengguna tidak lengkap di database"},
		{"nim kosong", "abc", &fakeSSO{userID: ptr(int64(1)), found: true, user: &entity.SSOUser{Nama: "B"}}, 400, "Data pengguna tidak lengkap di database"},
	}
	for _, tc := range cases {
		_, f := NewSSOUsecase(tc.repo, NewTokens(testSecret)).Generate(context.Background(), tc.code)
		if f == nil || f.Status != tc.status || f.Body["error"] != tc.msg {
			t.Errorf("%s: %+v", tc.name, f)
		}
	}
}

func TestGenerateJWTToken(t *testing.T) {
	cases := []struct {
		name    string
		user    entity.SSOUser
		lookups []string
		want    map[string]any
	}{
		{
			"praktikan memakai kelas jadwal",
			entity.SSOUser{Username: "u1", Nama: "Ani", Role: "praktikan", ClassCode: "A", Major: "TI"},
			[]string{"student"},
			map[string]any{"nim": "u1", "nama": "Ani", "kelas": "K1", "jurusan": "TI", "role": "praktikan"},
		},
		{
			"asisten, jadwal error diabaikan",
			entity.SSOUser{NIM: "n2", FullName: "Budi", Nama: "B", Role: "asisten", ClassCode: "B"},
			[]string{"assistant"},
			map[string]any{"nim": "n2", "nama": "Budi", "kelas": "K2", "jurusan": "SI", "role": "asisten"},
		},
		{
			"kordas tanpa lookup jadwal (T-22)",
			entity.SSOUser{NIM: "n3", FullName: "Cici", Role: "kordas", ClassCode: "C", Major: "MI"},
			nil,
			map[string]any{"nim": "n3", "nama": "Cici", "kelas": "C", "jurusan": "MI", "role": "kordas"},
		},
	}
	for _, tc := range cases {
		repo := &fakeSSO{
			userID: ptr(int64(9)), found: true, user: &tc.user,
			student: entity.ScheduleClass{ClassCode: "K1"}, assistant: entity.ScheduleClass{ClassCode: "K2", Major: "SI"},
		}
		uc := NewSSOUsecase(repo, NewTokens(testSecret))
		fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("WIB", 7*3600))
		uc.now = func() time.Time { return fixed }
		body, f := uc.Generate(context.Background(), "kode-1")
		if f != nil {
			t.Fatalf("%s: %+v", tc.name, f)
		}
		if !repo.gotNow.Equal(fixed) || repo.gotNow.Location() != time.UTC {
			t.Errorf("%s: now %v", tc.name, repo.gotNow)
		}
		if len(repo.deleted) != 1 || repo.deleted[0] != "kode-1" {
			t.Errorf("%s: handshake tidak dihapus: %v", tc.name, repo.deleted)
		}
		if len(repo.lookups) != len(tc.lookups) || (len(tc.lookups) > 0 && repo.lookups[0] != tc.lookups[0]) {
			t.Errorf("%s: lookup %v", tc.name, repo.lookups)
		}
		claims, err := NewTokens(testSecret).Verify(body["token"].(string))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		iat, _ := claims["iat"].(json.Number).Int64()
		exp, _ := claims["exp"].(json.Number).Int64()
		if exp-iat != int64((2 * time.Hour).Seconds()) {
			t.Errorf("%s: exp-iat %d", tc.name, exp-iat)
		}
		delete(claims, "iat")
		delete(claims, "exp")
		if got, want := string(mustJSON(t, claims)), string(mustJSON(t, tc.want)); got != want {
			t.Errorf("%s: klaim %s, ingin %s", tc.name, got, want)
		}
	}
}
