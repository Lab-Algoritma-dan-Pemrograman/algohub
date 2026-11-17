package database

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
)

type fakeOpener struct {
	calls []string
	err   error
}

func (f *fakeOpener) open(dsn string) (*gorm.DB, error) {
	f.calls = append(f.calls, dsn)
	if f.err != nil {
		return nil, f.err
	}
	return &gorm.DB{}, nil
}

func TestOpenUniqueSharesPoolForSameDSN(t *testing.T) {
	const dsn = "postgres://u@localhost:6543/db"
	f := &fakeOpener{}

	dbs, err := openUnique(config.DBConfig{Praktikum: dsn, Siakad: dsn, Elearning: dsn}, f.open)
	if err != nil {
		t.Fatalf("openUnique: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("opener dipanggil %d kali, ingin 1", len(f.calls))
	}
	if dbs.Praktikum == nil || dbs.Praktikum != dbs.Siakad || dbs.Praktikum != dbs.Elearning {
		t.Fatalf("ketiga pool harus pointer yang sama: %p %p %p", dbs.Praktikum, dbs.Siakad, dbs.Elearning)
	}
}

func TestOpenUniqueSkipsEmptyDSN(t *testing.T) {
	f := &fakeOpener{}

	dbs, err := openUnique(config.DBConfig{Praktikum: "postgres://u@localhost:6543/db"}, f.open)
	if err != nil {
		t.Fatalf("openUnique: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("opener dipanggil %d kali, ingin 1", len(f.calls))
	}
	if dbs.Praktikum == nil || dbs.Siakad != nil || dbs.Elearning != nil {
		t.Fatalf("hanya Praktikum yang boleh terisi: %p %p %p", dbs.Praktikum, dbs.Siakad, dbs.Elearning)
	}
}

func TestOpenUniqueSeparatesDistinctDSN(t *testing.T) {
	f := &fakeOpener{}

	dbs, err := openUnique(config.DBConfig{
		Praktikum: "postgres://u@localhost:6543/praktikum",
		Siakad:    "postgres://u@localhost:6543/siakad",
		Elearning: "postgres://u@localhost:6543/praktikum",
	}, f.open)
	if err != nil {
		t.Fatalf("openUnique: %v", err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("opener dipanggil %d kali, ingin 2", len(f.calls))
	}
	if dbs.Praktikum == dbs.Siakad {
		t.Fatal("DSN berbeda harus memakai pool berbeda")
	}
	if dbs.Praktikum != dbs.Elearning {
		t.Fatal("DSN sama harus memakai pool yang sama")
	}
}

func TestOpenUniquePropagatesError(t *testing.T) {
	want := errors.New("gagal konek")
	f := &fakeOpener{err: want}

	if _, err := openUnique(config.DBConfig{Praktikum: "postgres://u@localhost:6543/db"}, f.open); !errors.Is(err, want) {
		t.Fatalf("err = %v, ingin %v", err, want)
	}
}

func TestWithConnectTimeout(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
		want string
	}{
		{"url tanpa query", "postgres://u:p@host:6543/db", "postgres://u:p@host:6543/db?connect_timeout=15"},
		{"url dengan query", "postgresql://u:p@host:6543/db?sslmode=require", "postgresql://u:p@host:6543/db?connect_timeout=15&sslmode=require"},
		{"url sudah ada connect_timeout", "postgres://u@host/db?connect_timeout=5", "postgres://u@host/db?connect_timeout=5"},
		{"key=value", "host=localhost dbname=db sslmode=disable", "host=localhost dbname=db sslmode=disable connect_timeout=15"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := withConnectTimeout(tc.dsn); got != tc.want {
				t.Fatalf("withConnectTimeout(%q) = %q, ingin %q", tc.dsn, got, tc.want)
			}
		})
	}
}
