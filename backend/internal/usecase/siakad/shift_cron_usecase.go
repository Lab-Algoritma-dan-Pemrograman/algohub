package usecase

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

type ShiftCronUsecase struct {
	shifts repository.ShiftRepository
	push   repository.PushRepository
	sender PushSender
	now    func() time.Time
}

func NewShiftCronUsecase(shifts repository.ShiftRepository, push repository.PushRepository, sender PushSender) *ShiftCronUsecase {
	return &ShiftCronUsecase{shifts: shifts, push: push, sender: sender, now: time.Now}
}

// K-23: semua perhitungan di UTC seperti Node di Vercel, termasuk start_time yang sebenarnya WIB (T-12).
func (u *ShiftCronUsecase) Run(ctx context.Context) (map[string]any, *Failure) {
	now := u.now().UTC().Truncate(time.Millisecond)

	assignments, err := u.shifts.TodayAssignments(ctx, now.Format(time.DateOnly))
	if err != nil {
		return nil, fail(http.StatusInternalServerError, errText(err))
	}
	if len(assignments) == 0 {
		return map[string]any{"message": "No assignments today", "processed": 0}, nil
	}
	subs, err := u.push.List(ctx, nil)
	if err != nil || len(subs) == 0 {
		return map[string]any{"message": "No push subscriptions registered", "processed": 0}, nil
	}

	byUser := make(map[int64][]entity.PushSubscription)
	for _, s := range subs {
		if s.UserID != nil {
			byUser[*s.UserID] = append(byUser[*s.UserID], s)
		}
	}

	sent := 0
	for _, a := range assignments {
		if a.UserID == nil || *a.UserID == 0 || a.StartTime == "" {
			continue
		}
		userSubs := byUser[*a.UserID]
		if len(userSubs) == 0 {
			continue
		}
		title, body, ok := shiftReminder(a, now)
		if !ok {
			continue
		}
		payload, err := encodeJSON(pushPayload{Title: title, Body: body, URL: jadwalJagaURL})
		if err != nil {
			return nil, fail(http.StatusInternalServerError, err.Error())
		}
		for _, s := range userSubs {
			_, err := u.sender.Send(ctx, toWebpush(s), payload)
			switch {
			case err == nil:
				sent++
			case isExpired(err):
				_ = u.push.DeleteByEndpoint(ctx, s.Endpoint)
			}
		}
	}

	return map[string]any{"status": "ok", "timestamp": isoMillis(now), "notificationsSent": sent}, nil
}

func shiftReminder(a entity.ShiftAssignment, now time.Time) (title, body string, ok bool) {
	parts := strings.Split(a.StartTime, ":")
	if len(parts) < 2 {
		return "", "", false
	}
	hours, okH := parseLeadingInt(parts[0])
	minutes, okM := parseLeadingInt(parts[1])
	if !okH || !okM {
		return "", "", false
	}
	shift := time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, time.UTC)
	diff := int(math.Floor(float64(shift.Sub(now).Milliseconds()) / 60000))

	role := orDefault(a.UserRole, "asisten")
	quote := dailyQuote(role, now)
	start := a.StartTime
	if len(start) > 5 {
		start = start[:5]
	}
	taskRole := orDefault(a.TaskRole, "Petugas")
	activity := orDefault(a.ActivityName, "Praktikum")

	duty := taskRole + " di " + activity + " (Pukul " + start + " WIB)."
	closing := "\n\n✨ \"" + quote + "\""
	switch {
	case diff <= 30 && diff > 15:
		return "🚨 30 Menit Lagi! Jadwal Jaga Segera Dimulai",
			"Perhatian! 30 menit lagi kamu bertugas sebagai " + duty + " Segera menuju laboratorium!" + closing, true
	case diff <= 60 && diff > 45:
		return "⏳ 1 Jam Lagi! Jadwal Jaga", "Persiapkan diri! 1 jam lagi kamu bertugas sebagai " + duty + closing, true
	case diff <= 120 && diff > 105:
		return "⏰ 2 Jam Lagi! Jadwal Jaga", "Halo, 2 jam lagi kamu bertugas sebagai " + duty + closing, true
	}
	return "", "", false
}

// Meniru parseInt JavaScript: angka di awal string dipakai, sisa karakter diabaikan.
func parseLeadingInt(s string) (int, bool) {
	s = strings.TrimLeft(s, " \t\n\r")
	sign := 1
	if s != "" && (s[0] == '-' || s[0] == '+') {
		if s[0] == '-' {
			sign = -1
		}
		s = s[1:]
	}
	n, digits := 0, 0
	for digits < len(s) && s[digits] >= '0' && s[digits] <= '9' && digits < 9 {
		n = n*10 + int(s[digits]-'0')
		digits++
	}
	return sign * n, digits > 0
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
