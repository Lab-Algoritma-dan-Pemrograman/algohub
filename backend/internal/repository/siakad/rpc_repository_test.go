package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("gorm: %v", err)
	}
	return db, mock
}

// gorm menulis placeholder ? sebagai $1 untuk Postgres.
var metaSQL = strings.Replace(rpcMetaSQL, "?", "$1", 1)

func metaRows(setof, void bool, required int, args string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"proretset", "void", "required", "args"}).
		AddRow(setof, void, int64(required), []byte(args))
}

const userArgs = `[{"name":"p_user_id","type":"bigint"},{"name":"p_tags","type":"text[]"},{"name":"p_note","type":"text"}]`

func TestRPCCallCachesMetadata(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRPCRepository(db)
	params := map[string]json.RawMessage{"p_user_id": json.RawMessage(`5`), "p_tags": json.RawMessage(`["a","b"]`)}
	callSQL := `SELECT coalesce(json_agg(_r), '[]'::json) FROM json_to_record($1::json) AS _a("p_user_id" bigint, "p_tags" text[]) CROSS JOIN LATERAL public."get_x"("p_user_id" => _a."p_user_id", "p_tags" => _a."p_tags") AS _r`

	mock.ExpectQuery(metaSQL).WithArgs("get_x").WillReturnRows(metaRows(true, false, 1, userArgs))
	mock.ExpectQuery(callSQL).WithArgs(`{"p_tags":["a","b"],"p_user_id":5}`).
		WillReturnRows(sqlmock.NewRows([]string{"r"}).AddRow([]byte(`[{"id":1}]`)))
	mock.ExpectQuery(callSQL).WithArgs(`{"p_tags":["a","b"],"p_user_id":5}`).
		WillReturnRows(sqlmock.NewRows([]string{"r"}).AddRow([]byte(`[]`)))

	got, err := repo.Call(context.Background(), "get_x", params)
	if err != nil || string(got) != `[{"id":1}]` {
		t.Fatalf("Call 1 = %s, %v", got, err)
	}
	got, err = repo.Call(context.Background(), "get_x", params)
	if err != nil || string(got) != `[]` {
		t.Fatalf("Call 2 = %s, %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRPCCallNullAndNoArgs(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRPCRepository(db)

	mock.ExpectQuery(metaSQL).WithArgs("get_public_settings").WillReturnRows(metaRows(false, false, 0, `[]`))
	mock.ExpectQuery(`SELECT to_json(_r) FROM public."get_public_settings"() AS _r`).
		WillReturnRows(sqlmock.NewRows([]string{"r"}).AddRow(nil))

	got, err := repo.Call(context.Background(), "get_public_settings", nil)
	if err != nil || string(got) != "null" {
		t.Fatalf("Call = %s, %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRPCCallUnknownKey(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRPCRepository(db)
	mock.ExpectQuery(metaSQL).WithArgs("get_x").WillReturnRows(metaRows(true, false, 1, userArgs))

	cases := []map[string]json.RawMessage{
		{"p_user_id": json.RawMessage(`1`), "p_zzz": json.RawMessage(`1`), "p_aaa": json.RawMessage(`2`)},
		{"p_note": json.RawMessage(`"x"`)},
	}
	wants := []string{
		"Could not find the function public.get_x(p_aaa, p_user_id, p_zzz) in the schema cache",
		"Could not find the function public.get_x(p_note) in the schema cache",
	}
	for i, params := range cases {
		_, err := repo.Call(context.Background(), "get_x", params)
		var nf *FunctionNotFoundError
		if !errors.As(err, &nf) || err.Error() != wants[i] {
			t.Errorf("kasus %d: err = %v, ingin %q", i, err, wants[i])
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRPCCallMissingFunctionNotCached(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRPCRepository(db)
	for range 2 {
		mock.ExpectQuery(metaSQL).WithArgs("hilang").WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c", "d"}))
	}
	for range 2 {
		var nf *FunctionNotFoundError
		if _, err := repo.Call(context.Background(), "hilang", nil); !errors.As(err, &nf) {
			t.Fatalf("err = %v, ingin FunctionNotFoundError", err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRPCCallPassesPgError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRPCRepository(db)
	pgErr := &pgconn.PgError{Code: "P0001", Message: "Akses ditolak"}

	mock.ExpectQuery(metaSQL).WithArgs("do_x").WillReturnRows(metaRows(false, true, 0, `[{"name":"p_note","type":"text"}]`))
	mock.ExpectQuery(`SELECT NULL::json FROM json_to_record($1::json) AS _a("p_note" text) CROSS JOIN LATERAL public."do_x"("p_note" => _a."p_note") AS _r`).
		WithArgs(`{"p_note":"x"}`).WillReturnError(pgErr)

	_, err := repo.Call(context.Background(), "do_x", map[string]json.RawMessage{"p_note": json.RawMessage(`"x"`)})
	var got *pgconn.PgError
	if !errors.As(err, &got) || got.Message != "Akses ditolak" {
		t.Fatalf("err = %v, ingin PgError utuh", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBuildCallSQL(t *testing.T) {
	args := []rpcArg{{"p_a", "bigint"}, {"p_b", "timestamp with time zone"}}
	both := map[string]json.RawMessage{"p_b": nil, "p_a": nil}
	cases := []struct {
		name   string
		fn     string
		o      rpcOverload
		params map[string]json.RawMessage
		want   string
	}{
		{"setof", "f", rpcOverload{setof: true, args: args}, both,
			`SELECT coalesce(json_agg(_r), '[]'::json) FROM json_to_record(?::json) AS _a("p_a" bigint, "p_b" timestamp with time zone) CROSS JOIN LATERAL public."f"("p_a" => _a."p_a", "p_b" => _a."p_b") AS _r`},
		{"scalar sebagian", "f", rpcOverload{args: args}, map[string]json.RawMessage{"p_b": nil},
			`SELECT to_json(_r) FROM json_to_record(?::json) AS _a("p_b" timestamp with time zone) CROSS JOIN LATERAL public."f"("p_b" => _a."p_b") AS _r`},
		{"void", "f", rpcOverload{void: true, args: args}, map[string]json.RawMessage{"p_a": nil},
			`SELECT NULL::json FROM json_to_record(?::json) AS _a("p_a" bigint) CROSS JOIN LATERAL public."f"("p_a" => _a."p_a") AS _r`},
		{"tanpa argumen", "f", rpcOverload{setof: true}, nil,
			`SELECT coalesce(json_agg(_r), '[]'::json) FROM public."f"() AS _r`},
		{"nama di-quote", `x"; DROP TABLE users; --`, rpcOverload{args: []rpcArg{{`p"q`, "text"}}}, map[string]json.RawMessage{`p"q`: nil},
			`SELECT to_json(_r) FROM json_to_record(?::json) AS _a("p""q" text) CROSS JOIN LATERAL public."x""; DROP TABLE users; --"("p""q" => _a."p""q") AS _r`},
	}
	for _, tc := range cases {
		if got := buildCallSQL(tc.fn, tc.o, tc.params); got != tc.want {
			t.Errorf("%s:\n got  %s\n want %s", tc.name, got, tc.want)
		}
	}
}
