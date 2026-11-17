package usecase

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

type fakeShifts struct {
	rows  []entity.ShiftAssignment
	err   error
	dates []string
}

func (f *fakeShifts) TodayAssignments(_ context.Context, date string) ([]entity.ShiftAssignment, error) {
	f.dates = append(f.dates, date)
	return f.rows, f.err
}

func newShiftCron(rows []entity.ShiftAssignment, subs []entity.PushSubscription, sender *fakeSender, now time.Time) (*ShiftCronUsecase, *fakeShifts, *fakePush) {
	shifts := &fakeShifts{rows: rows}
	push := &fakePush{subs: subs}
	uc := NewShiftCronUsecase(shifts, push, sender)
	uc.now = func() time.Time { return now }
	return uc, shifts, push
}

func TestShiftCronWindows(t *testing.T) {
	// 17 Nov 2026 = hari ke-321, jadi kutipan = dasar ke-11 + penguat ke-3 (indeks 2).
	now := time.Date(2026, 11, 17, 7, 0, 0, 500_000_000, time.UTC)
	asistenQuote := asistenBaseQuotes[10] + " " + asistenEnhancers[2]
	praktikanQuote := praktikanBaseQuotes[10] + " " + praktikanEnhancers[2]

	cases := []struct {
		name  string
		row   entity.ShiftAssignment
		title string
		body  string
	}{
		{
			"30 menit", entity.ShiftAssignment{UserID: uid(7), StartTime: "07:30:00", UserRole: "Koordinator", TaskRole: "PJ", ActivityName: "Praktikum AP"},
			"🚨 30 Menit Lagi! Jadwal Jaga Segera Dimulai",
			"Perhatian! 30 menit lagi kamu bertugas sebagai PJ di Praktikum AP (Pukul 07:30 WIB). Segera menuju laboratorium!\n\n✨ \"" + asistenQuote + "\"",
		},
		{
			"1 jam, default teks", entity.ShiftAssignment{UserID: uid(7), StartTime: "08:00:00"},
			"⏳ 1 Jam Lagi! Jadwal Jaga",
			"Persiapkan diri! 1 jam lagi kamu bertugas sebagai Petugas di Praktikum (Pukul 08:00 WIB).\n\n✨ \"" + asistenQuote + "\"",
		},
		{
			"2 jam, role praktikan", entity.ShiftAssignment{UserID: uid(7), StartTime: "08:47", UserRole: "mahasiswa", TaskRole: "Jaga", ActivityName: "Ujian"},
			"⏰ 2 Jam Lagi! Jadwal Jaga",
			"Halo, 2 jam lagi kamu bertugas sebagai Jaga di Ujian (Pukul 08:47 WIB).\n\n✨ \"" + praktikanQuote + "\"",
		},
	}
	for _, tc := range cases {
		sender := &fakeSender{}
		uc, shifts, _ := newShiftCron([]entity.ShiftAssignment{tc.row}, pushSubs(), sender, now)
		body, f := uc.Run(context.Background())
		if f != nil {
			t.Fatalf("%s: %+v", tc.name, f)
		}
		if !mapsEqual(body, map[string]any{"status": "ok", "timestamp": "2026-11-17T07:00:00.500Z", "notificationsSent": 2}) {
			t.Errorf("%s: body = %v", tc.name, body)
		}
		if shifts.dates[0] != "2026-11-17" {
			t.Errorf("%s: tanggal = %v", tc.name, shifts.dates)
		}
		if len(sender.sent) != 2 || sender.sent[0].endpoint != "https://e/7a" || sender.sent[1].endpoint != "https://e/7b" {
			t.Fatalf("%s: terkirim = %+v", tc.name, sender.sent)
		}
		p := sender.sent[0].payload
		if p["title"] != tc.title || p["body"] != tc.body || p["url"] != "/jadwal-jaga" {
			t.Errorf("%s: payload\n%q\n%q\ningin\n%q\n%q", tc.name, p["title"], p["body"], tc.title, tc.body)
		}
	}
}

func TestShiftCronOutsideWindow(t *testing.T) {
	now := time.Date(2026, 11, 17, 7, 0, 0, 0, time.UTC)
	rows := []entity.ShiftAssignment{
		{UserID: uid(7), StartTime: "07:15:00"},
		{UserID: uid(7), StartTime: "07:45:00"},
		{UserID: uid(7), StartTime: "08:45:00"},
		{UserID: uid(7), StartTime: "09:01:00"},
		{UserID: uid(7), StartTime: "06:30:00"},
		{UserID: uid(7), StartTime: ""},
		{UserID: nil, StartTime: "07:30:00"},
		{UserID: uid(0), StartTime: "07:30:00"},
		{UserID: uid(99), StartTime: "07:30:00"},
		{UserID: uid(7), StartTime: "xx:30"},
	}
	sender := &fakeSender{}
	uc, _, _ := newShiftCron(rows, pushSubs(), sender, now)
	body, f := uc.Run(context.Background())
	if f != nil || body["notificationsSent"] != 0 || len(sender.sent) != 0 {
		t.Fatalf("di luar jendela: %v %+v %+v", body, f, sender.sent)
	}
}

func TestShiftCronCleansExpired(t *testing.T) {
	now := time.Date(2026, 11, 17, 7, 0, 0, 0, time.UTC)
	sender := &fakeSender{status: map[string]int{"https://e/7a": http.StatusGone, "https://e/7b": http.StatusNotFound, "https://e/8": http.StatusInternalServerError}}
	rows := []entity.ShiftAssignment{{UserID: uid(7), StartTime: "07:20:00"}, {UserID: uid(8), StartTime: "07:20:00"}}
	uc, _, push := newShiftCron(rows, pushSubs(), sender, now)
	body, f := uc.Run(context.Background())
	if f != nil || body["notificationsSent"] != 0 {
		t.Fatalf("%v %+v", body, f)
	}
	if !slices.Equal(push.deleted1, []string{"https://e/7a", "https://e/7b"}) {
		t.Fatalf("dihapus = %v", push.deleted1)
	}
}

func TestShiftCronEmpty(t *testing.T) {
	now := time.Date(2026, 11, 17, 7, 0, 0, 0, time.UTC)
	uc, _, _ := newShiftCron(nil, pushSubs(), &fakeSender{}, now)
	if body, f := uc.Run(context.Background()); f != nil || !mapsEqual(body, map[string]any{"message": "No assignments today", "processed": 0}) {
		t.Fatalf("tanpa jadwal: %v %+v", body, f)
	}

	uc, _, _ = newShiftCron([]entity.ShiftAssignment{{UserID: uid(7), StartTime: "07:30"}}, nil, &fakeSender{}, now)
	if body, f := uc.Run(context.Background()); f != nil || !mapsEqual(body, map[string]any{"message": "No push subscriptions registered", "processed": 0}) {
		t.Fatalf("tanpa subscription: %v %+v", body, f)
	}

	uc, shifts, _ := newShiftCron(nil, nil, &fakeSender{}, now)
	shifts.err = errors.New("relation does not exist")
	if _, f := uc.Run(context.Background()); f == nil || f.Status != 500 || f.Body["error"] != "relation does not exist" {
		t.Fatalf("error query: %+v", f)
	}
}

func TestDailyQuote(t *testing.T) {
	jan1 := time.Date(2026, 1, 1, 23, 0, 0, 0, time.UTC)
	if got := dailyQuote("asisten", jan1); got != asistenBaseQuotes[0] {
		t.Errorf("1 Jan = %q", got)
	}
	if got := dailyQuote("praktikan", jan1); got != praktikanBaseQuotes[0] {
		t.Errorf("1 Jan praktikan = %q", got)
	}
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if got := dailyQuote("KOORDINATOR", feb1); got != asistenBaseQuotes[0]+" "+asistenEnhancers[1] {
		t.Errorf("1 Feb = %q", got)
	}
	if len(asistenBaseQuotes) != 31 || len(praktikanBaseQuotes) != 31 || len(asistenEnhancers) != 8 || len(praktikanEnhancers) != 8 {
		t.Fatal("jumlah kutipan berubah")
	}
	if got := get365Quote(366, false); got != praktikanBaseQuotes[24]+" "+praktikanEnhancers[3] {
		t.Errorf("hari ke-366 = %q", got)
	}
}

type fakePinger struct {
	err   error
	calls int
}

func (f *fakePinger) Ping(context.Context) error {
	f.calls++
	return f.err
}

func TestKeepAlive(t *testing.T) {
	siakad, elearning := &fakePinger{}, &fakePinger{}
	uc := NewKeepAliveUsecase(siakad, elearning)
	uc.now = func() time.Time { return time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC) }
	body, f := uc.Run(context.Background())
	want := map[string]any{"success": true, "message": "Supabase pinged successfully", "timestamp": "2026-11-17T00:00:00.000Z"}
	if f != nil || !mapsEqual(body, want) || siakad.calls != 1 || elearning.calls != 1 {
		t.Fatalf("%v %+v", body, f)
	}

	siakad.err = errors.New("siakad paused")
	_, f = uc.Run(context.Background())
	if f == nil || f.Status != 500 || f.Body["error"] != "siakad paused" || elearning.calls != 2 {
		t.Fatalf("error ping: %+v, elearning dipanggil %d kali", f, elearning.calls)
	}
}
