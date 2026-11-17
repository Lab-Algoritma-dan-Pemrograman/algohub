package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/webpush"
)

type fakePush struct {
	subs     []entity.PushSubscription
	err      error
	filters  []any
	deleted  [][]string
	deleted1 []string
}

func (f *fakePush) List(_ context.Context, userFilter any) ([]entity.PushSubscription, error) {
	f.filters = append(f.filters, userFilter)
	if f.err != nil {
		return nil, f.err
	}
	if userFilter == nil {
		return f.subs, nil
	}
	var out []entity.PushSubscription
	for _, s := range f.subs {
		if id, ok := userFilter.(int64); ok && s.UserID != nil && *s.UserID == id {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakePush) DeleteByEndpoints(_ context.Context, endpoints []string) error {
	f.deleted = append(f.deleted, endpoints)
	return nil
}

func (f *fakePush) DeleteByEndpoint(_ context.Context, endpoint string) error {
	f.deleted1 = append(f.deleted1, endpoint)
	return nil
}

type sentPush struct {
	endpoint string
	payload  map[string]string
}

// status per endpoint; endpoint yang tidak terdaftar dianggap sukses 201.
type fakeSender struct {
	mu     sync.Mutex
	status map[string]int
	sent   []sentPush
}

func (f *fakeSender) Send(_ context.Context, sub webpush.Subscription, payload []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var p map[string]string
	_ = json.Unmarshal(payload, &p)
	f.sent = append(f.sent, sentPush{sub.Endpoint, p})
	if code, ok := f.status[sub.Endpoint]; ok {
		return code, &webpush.StatusError{Code: code}
	}
	return http.StatusCreated, nil
}

func uid(n int64) *int64 { return &n }

func pushSubs() []entity.PushSubscription {
	return []entity.PushSubscription{
		{UserID: uid(7), Endpoint: "https://e/7a"},
		{UserID: uid(7), Endpoint: "https://e/7b"},
		{UserID: uid(8), Endpoint: "https://e/8"},
	}
}

func TestPushAuthorization(t *testing.T) {
	uc := NewPushUsecase(&fakePush{subs: pushSubs()}, &fakeSender{}, NewTokens(testSecret))
	req := dto.SendPushRequest{Title: "T", Body: "B"}
	cases := []struct {
		name     string
		fromCron bool
		token    string
		status   int
		msg      string
	}{
		{"cron secret", true, "", 0, ""},
		{"tanpa token", false, "", 401, "Unauthorized: Missing session token"},
		{"token rusak", false, "rusak", 401, "Unauthorized: Invalid or expired token"},
		{"praktikan", false, sign(t, map[string]any{"role": "mahasiswa"}), 403, "Forbidden: Hanya staf yang boleh mengirim notifikasi"},
		{"korda bukan staf lama", false, sign(t, map[string]any{"role": "korda"}), 403, "Forbidden: Hanya staf yang boleh mengirim notifikasi"},
		{"role berspasi", false, sign(t, map[string]any{"role": " asisten"}), 403, "Forbidden: Hanya staf yang boleh mengirim notifikasi"},
		{"Kordas", false, sign(t, map[string]any{"role": "Kordas"}), 0, ""},
		{"admin", false, sign(t, map[string]any{"role": "admin"}), 0, ""},
	}
	for _, tc := range cases {
		_, f := uc.Send(context.Background(), tc.fromCron, tc.token, req)
		switch {
		case tc.status == 0 && f != nil:
			t.Errorf("%s: harus lolos, dapat %+v", tc.name, f)
		case tc.status != 0 && (f == nil || f.Status != tc.status || f.Body["error"] != tc.msg):
			t.Errorf("%s: %+v, ingin %d %q", tc.name, f, tc.status, tc.msg)
		}
	}
}

func TestPushSend(t *testing.T) {
	repo := &fakePush{subs: pushSubs()}
	sender := &fakeSender{status: map[string]int{"https://e/7b": http.StatusGone}}
	uc := NewPushUsecase(repo, sender, NewTokens(testSecret))

	if _, f := uc.Send(context.Background(), true, "", dto.SendPushRequest{Title: "T"}); f == nil ||
		f.Status != 400 || f.Body["error"] != "Title and body are required" {
		t.Fatalf("body kosong: %+v", f)
	}

	body, f := uc.Send(context.Background(), true, "", dto.SendPushRequest{TargetUserID: "7", Title: "T", Body: "B"})
	if f != nil {
		t.Fatal(f)
	}
	want := map[string]any{"message": "Web Push processing complete", "sent": 1, "total": 2, "cleaned": 1}
	if !mapsEqual(body, want) {
		t.Fatalf("body = %v", body)
	}
	if len(repo.filters) != 1 || repo.filters[0] != int64(7) {
		t.Fatalf("filter = %#v", repo.filters)
	}
	if len(repo.deleted) != 1 || !slices.Equal(repo.deleted[0], []string{"https://e/7b"}) {
		t.Fatalf("dihapus = %v", repo.deleted)
	}
	for _, s := range sender.sent {
		if s.payload["title"] != "T" || s.payload["body"] != "B" || s.payload["url"] != "/jadwal-jaga" {
			t.Fatalf("payload = %v", s.payload)
		}
	}
}

func TestPushTargets(t *testing.T) {
	cases := []struct {
		target any
		want   any
	}{
		{nil, nil},
		{"", nil},
		{float64(0), nil},
		{float64(8), int64(8)},
		{" 8 ", int64(8)},
		{"abc", "abc"},
		{float64(1.5), "1.5"},
	}
	for _, tc := range cases {
		if got := pushUserFilter(tc.target); got != tc.want {
			t.Errorf("pushUserFilter(%#v) = %#v, ingin %#v", tc.target, got, tc.want)
		}
	}
}

func TestPushEmptyAndError(t *testing.T) {
	uc := NewPushUsecase(&fakePush{}, &fakeSender{}, NewTokens(testSecret))
	body, f := uc.Send(context.Background(), true, "", dto.SendPushRequest{Title: "T", Body: "B", URL: "/x"})
	if f != nil || !mapsEqual(body, map[string]any{"message": "No target subscriptions found", "sent": 0}) {
		t.Fatalf("kosong: %v %+v", body, f)
	}

	pgErr := &pgconn.PgError{Message: `invalid input syntax for type bigint: "abc"`}
	uc = NewPushUsecase(&fakePush{err: pgErr}, &fakeSender{}, NewTokens(testSecret))
	_, f = uc.Send(context.Background(), true, "", dto.SendPushRequest{TargetUserID: "abc", Title: "T", Body: "B"})
	if f == nil || f.Status != 500 || f.Body["error"] != pgErr.Message {
		t.Fatalf("error query: %+v", f)
	}
}

type fakeNotify struct {
	settings entity.WASettings
	err      error
	phones   map[int64]string
	lookups  []int64
}

func (f *fakeNotify) WASettings(context.Context) (entity.WASettings, error) { return f.settings, f.err }

func (f *fakeNotify) UserPhone(_ context.Context, id int64) (string, error) {
	f.lookups = append(f.lookups, id)
	return f.phones[id], nil
}

type waCall struct{ provider, token, phone, message string }

type fakeWA struct {
	calls []waCall
	err   error
}

func (f *fakeWA) Send(_ context.Context, provider, token, phone, message string) (string, error) {
	f.calls = append(f.calls, waCall{provider, token, phone, message})
	return `{"status":true}`, f.err
}

func TestWASend(t *testing.T) {
	off := false
	on := true
	staff := sign(t, map[string]any{"role": "koordinator"})
	cases := []struct {
		name     string
		notify   *fakeNotify
		envToken string
		token    string
		req      dto.SendWARequest
		status   int
		body     map[string]any
		call     *waCall
	}{
		{
			name: "non-staf", notify: &fakeNotify{}, token: sign(t, map[string]any{"role": "praktikan"}),
			req: dto.SendWARequest{Message: "Hi"}, status: 403,
			body: map[string]any{"error": "Forbidden: Hanya staf yang boleh mengirim WhatsApp"},
		},
		{
			name: "pesan kosong", notify: &fakeNotify{}, token: staff, status: 400,
			body: map[string]any{"error": "Message content is required"},
		},
		{
			name: "disabled", notify: &fakeNotify{settings: entity.WASettings{Token: "t", AutoNotifyEnabled: &off}}, token: staff,
			req: dto.SendWARequest{TargetPhone: "0812", Message: "Hi"}, status: 200,
			body: map[string]any{"message": "WhatsApp auto-notify is disabled in settings", "sent": false},
		},
		{
			name: "token kosong", notify: &fakeNotify{err: errors.New("putus")}, token: staff,
			req: dto.SendWARequest{TargetPhone: "0812", Message: "Hi"}, status: 400,
			body: map[string]any{"error": "WA Gateway API Token belum dikonfigurasi. Silakan isi API Token di menu Pengaturan.", "sent": false},
		},
		{
			name: "nomor tidak ada", notify: &fakeNotify{settings: entity.WASettings{Token: "t"}}, token: staff,
			req: dto.SendWARequest{TargetUserID: float64(9), Message: "Hi"}, status: 400,
			body: map[string]any{"error": "Nomor WhatsApp tujuan tidak ditemukan.", "sent": false},
		},
		{
			name: "nomor dari user, token env, provider default", notify: &fakeNotify{phones: map[int64]string{7: "0812-3456 789"}},
			envToken: "env-tok", token: staff, req: dto.SendWARequest{TargetUserID: "7", Message: "Hi"}, status: 200,
			body: map[string]any{"status": "ok", "provider": "fonnte", "recipient": "628123456789", "response": `{"status":true}`, "sent": true},
			call: &waCall{"fonnte", "env-tok", "628123456789", "Hi"},
		},
		{
			name: "settings menang, nomor +62", notify: &fakeNotify{settings: entity.WASettings{Provider: "wablas", Token: "db-tok", AutoNotifyEnabled: &on}},
			envToken: "env-tok", token: "", req: dto.SendWARequest{TargetPhone: "+62 812", Message: "Hi"}, status: 200,
			body: map[string]any{"status": "ok", "provider": "wablas", "recipient": "62812", "response": `{"status":true}`, "sent": true},
			call: &waCall{"wablas", "db-tok", "62812", "Hi"},
		},
	}
	for _, tc := range cases {
		wa := &fakeWA{}
		uc := NewWAUsecase(tc.notify, wa, NewTokens(testSecret), tc.envToken)
		body, f := uc.Send(context.Background(), tc.token == "", tc.token, tc.req)
		status := 200
		if f != nil {
			status, body = f.Status, f.Body
		}
		if status != tc.status || !mapsEqual(body, tc.body) {
			t.Errorf("%s: %d %v", tc.name, status, body)
		}
		if tc.call == nil && len(wa.calls) != 0 || tc.call != nil && (len(wa.calls) != 1 || wa.calls[0] != *tc.call) {
			t.Errorf("%s: kirim = %+v", tc.name, wa.calls)
		}
	}
}

func TestWASendError(t *testing.T) {
	wa := &fakeWA{err: errors.New("dial tcp: timeout")}
	uc := NewWAUsecase(&fakeNotify{}, wa, NewTokens(testSecret), "tok")
	_, f := uc.Send(context.Background(), true, "", dto.SendWARequest{TargetPhone: "0812", Message: "Hi"})
	if f == nil || f.Status != 500 || !mapsEqual(f.Body, map[string]any{"error": "dial tcp: timeout", "sent": false}) {
		t.Fatalf("%+v", f)
	}
}

func mapsEqual(a, b map[string]any) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return bytes.Equal(ja, jb)
}
