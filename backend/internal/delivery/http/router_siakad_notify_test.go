package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	cronhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/cron"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/webpush"
)

const cronSecret = "cron-uji"

type fakeSiakadPush struct{ deleted []string }

func (f *fakeSiakadPush) List(context.Context, any) ([]entity.PushSubscription, error) {
	id := int64(7)
	return []entity.PushSubscription{{UserID: &id, Endpoint: "https://e/ok"}, {UserID: &id, Endpoint: "https://e/gone"}}, nil
}

func (f *fakeSiakadPush) DeleteByEndpoints(_ context.Context, endpoints []string) error {
	f.deleted = append(f.deleted, endpoints...)
	return nil
}

func (f *fakeSiakadPush) DeleteByEndpoint(_ context.Context, endpoint string) error {
	f.deleted = append(f.deleted, endpoint)
	return nil
}

type fakeSiakadPusher struct{}

func (fakeSiakadPusher) Send(_ context.Context, sub webpush.Subscription, _ []byte) (int, error) {
	if sub.Endpoint == "https://e/gone" {
		return http.StatusGone, &webpush.StatusError{Code: http.StatusGone}
	}
	return http.StatusCreated, nil
}

type fakeSiakadNotify struct{}

func (fakeSiakadNotify) WASettings(context.Context) (entity.WASettings, error) {
	return entity.WASettings{Provider: "fonnte", Token: "tok"}, nil
}

func (fakeSiakadNotify) UserPhone(context.Context, int64) (string, error) { return "08123", nil }

type fakeSiakadWA struct{}

func (fakeSiakadWA) Send(context.Context, string, string, string, string) (string, error) {
	return "ok", nil
}

type fakeSiakadShifts struct{}

func (fakeSiakadShifts) TodayAssignments(context.Context, string) ([]entity.ShiftAssignment, error) {
	return nil, nil
}

type fakeSiakadPinger struct {
	err   error
	calls int
}

func (f *fakeSiakadPinger) Ping(context.Context) error {
	f.calls++
	return f.err
}

type notifyRig struct {
	r         *gin.Engine
	push      *fakeSiakadPush
	siakad    *fakeSiakadPinger
	elearning *fakeSiakadPinger
}

func newNotifyRouter(t *testing.T) *notifyRig {
	t.Helper()
	tokens := usecase.NewTokens(siakadSecret)
	rig := &notifyRig{push: &fakeSiakadPush{}, siakad: &fakeSiakadPinger{}, elearning: &fakeSiakadPinger{}}
	rig.r = newTestRouter(t)
	RegisterSiakad(rig.r, &Siakad{Handlers: SiakadHandlers{
		Notify: handler.NewNotifyHandler(
			usecase.NewPushUsecase(rig.push, fakeSiakadPusher{}, tokens),
			usecase.NewWAUsecase(fakeSiakadNotify{}, fakeSiakadWA{}, tokens, ""),
			cronSecret,
		),
		Cron: cronhandler.NewCronHandler(cronSecret,
			usecase.NewShiftCronUsecase(fakeSiakadShifts{}, rig.push, fakeSiakadPusher{}),
			usecase.NewKeepAliveUsecase(rig.siakad, rig.elearning),
		),
	}})
	return rig
}

func TestSiakadSendPush(t *testing.T) {
	rig := newNotifyRouter(t)
	body := `{"title":"T","body":"B","targetUserId":7}`
	staff := map[string]string{"Cookie": "siakad_token=" + siakadToken(t, map[string]any{"role": "asisten"})}

	cases := []struct {
		name   string
		header map[string]string
		status int
		want   string
	}{
		{"x-cron-secret", map[string]string{"X-Cron-Secret": cronSecret}, 200, `{"cleaned":1,"message":"Web Push processing complete","sent":1,"total":2}`},
		{"cookie staf", staff, 200, `{"cleaned":1,"message":"Web Push processing complete","sent":1,"total":2}`},
		{"bearer cron bukan token sesi", map[string]string{"Authorization": "Bearer " + cronSecret}, 401, `{"error":"Unauthorized: Invalid or expired token"}`},
		{"x-cron-secret salah", map[string]string{"X-Cron-Secret": "salah"}, 401, `{"error":"Unauthorized: Missing session token"}`},
		{"non-staf", map[string]string{"Authorization": "Bearer " + siakadToken(t, map[string]any{"role": "praktikan"})}, 403, `{"error":"Forbidden: Hanya staf yang boleh mengirim notifikasi"}`},
	}
	for _, tc := range cases {
		rec := send(rig.r, http.MethodPost, "/api/siakad/send-push", body, tc.header)
		if rec.Code != tc.status || rec.Body.String() != tc.want {
			t.Errorf("%s: %d %s", tc.name, rec.Code, rec.Body)
		}
	}
	if len(rig.push.deleted) != 2 || rig.push.deleted[0] != "https://e/gone" {
		t.Fatalf("dihapus = %v", rig.push.deleted)
	}
}

func TestSiakadSendWA(t *testing.T) {
	rig := newNotifyRouter(t)
	rec := send(rig.r, http.MethodPost, "/api/siakad/send-wa", `{"targetUserId":"7","message":"Hi"}`, map[string]string{"X-Cron-Secret": cronSecret})
	want := `{"provider":"fonnte","recipient":"628123","response":"ok","sent":true,"status":"ok"}`
	if rec.Code != http.StatusOK || rec.Body.String() != want {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = send(rig.r, http.MethodPost, "/api/siakad/send-wa", `{"message":"Hi"}`, nil)
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Missing session token"}` {
		t.Fatalf("tanpa token: %d %s", rec.Code, rec.Body)
	}
}

func TestSiakadNotifyMethods(t *testing.T) {
	rig := newNotifyRouter(t)
	for _, path := range []string{"/api/siakad/send-push", "/api/siakad/send-wa"} {
		if rec := send(rig.r, http.MethodOptions, path, "", nil); rec.Code != http.StatusOK || rec.Body.Len() != 0 {
			t.Errorf("OPTIONS %s: %d %q", path, rec.Code, rec.Body)
		}
		if rec := send(rig.r, http.MethodGet, path, "", nil); rec.Code != http.StatusMethodNotAllowed || rec.Body.String() != `{"error":"Method not allowed"}` {
			t.Errorf("GET %s: %d %s", path, rec.Code, rec.Body)
		}
	}
	rec := serve(rig.r, http.MethodOptions, "/api/siakad/send-push", allowedOrigin, map[string]string{"Access-Control-Request-Method": "POST"})
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != allowedOrigin {
		t.Fatalf("preflight: %d %v", rec.Code, rec.Header())
	}
}

func TestSiakadCheckShifts(t *testing.T) {
	rig := newNotifyRouter(t)
	rec := send(rig.r, http.MethodGet, "/api/cron/siakad/check-shifts", "", nil)
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Access restricted to Vercel Cron only"}` {
		t.Fatalf("tanpa secret: %d %s", rec.Code, rec.Body)
	}
	rec = send(rig.r, http.MethodGet, "/api/cron/siakad/check-shifts", "", map[string]string{"Authorization": "Bearer " + cronSecret})
	if rec.Code != http.StatusOK || rec.Body.String() != `{"message":"No assignments today","processed":0}` {
		t.Fatalf("bearer: %d %s", rec.Code, rec.Body)
	}
	if rec := send(rig.r, http.MethodPost, "/api/cron/siakad/check-shifts", "", map[string]string{"Authorization": "Bearer " + cronSecret}); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST: %d", rec.Code)
	}
}

func TestSiakadKeepAlive(t *testing.T) {
	rig := newNotifyRouter(t)
	bearer := map[string]string{"Authorization": "Bearer " + cronSecret}

	rec := send(rig.r, http.MethodGet, "/api/cron/keep-alive", "", bearer)
	if rec.Code != http.StatusOK || rig.siakad.calls != 1 || rig.elearning.calls != 1 {
		t.Fatalf("bearer: %d %s, ping siakad %d elearning %d", rec.Code, rec.Body, rig.siakad.calls, rig.elearning.calls)
	}

	for _, h := range []map[string]string{nil, {"x-vercel-cron": "1"}, {"Authorization": "Bearer salah"}} {
		rec = send(rig.r, http.MethodGet, "/api/cron/keep-alive", "", h)
		if rec.Code != http.StatusUnauthorized || rec.Body.String() != `{"error":"Unauthorized: Access restricted to Vercel Cron only"}` {
			t.Errorf("header %v: %d %s", h, rec.Code, rec.Body)
		}
	}
	if rig.siakad.calls != 1 {
		t.Fatal("DB tidak boleh di-ping tanpa secret")
	}

	rig.elearning.err = errors.New("project paused")
	rec = send(rig.r, http.MethodGet, "/api/cron/keep-alive", "", bearer)
	if rec.Code != http.StatusInternalServerError || rec.Body.String() != `{"error":"project paused"}` {
		t.Fatalf("error ping: %d %s", rec.Code, rec.Body)
	}

	rec = send(rig.r, http.MethodPost, "/api/cron/keep-alive", "", bearer)
	if rec.Code != http.StatusMethodNotAllowed || rec.Body.String() != `{"error":"Method not allowed"}` {
		t.Fatalf("POST: %d %s", rec.Code, rec.Body)
	}
}
