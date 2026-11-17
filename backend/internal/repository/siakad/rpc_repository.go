package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type RPCRepository interface {
	Call(ctx context.Context, fn string, params map[string]json.RawMessage) (json.RawMessage, error)
}

// Pesan meniru PGRST202 supaya frontend lama menerima error yang sama dengan supabase.rpc.
type FunctionNotFoundError struct {
	Fn   string
	Keys []string
}

func (e *FunctionNotFoundError) Error() string {
	return fmt.Sprintf("Could not find the function public.%s(%s) in the schema cache", e.Fn, strings.Join(e.Keys, ", "))
}

// Hanya argumen IN/INOUT/VARIADIC yang bisa dikirim; kolom RETURNS TABLE (mode 't') dan OUT dibuang.
const rpcMetaSQL = `SELECT p.proretset, p.prorettype = 'pg_catalog.void'::regtype, p.pronargs - p.pronargdefaults, coalesce((SELECT json_agg(json_build_object('name', a.name, 'type', format_type(a.type, NULL)) ORDER BY a.ord) FROM unnest(coalesce(p.proallargtypes, p.proargtypes::oid[]), p.proargnames, coalesce(p.proargmodes, array_fill('i'::"char", ARRAY[cardinality(coalesce(p.proallargtypes, p.proargtypes::oid[]))]))) WITH ORDINALITY AS a(type, name, mode, ord) WHERE a.mode IN ('i', 'b', 'v')), '[]'::json) FROM pg_catalog.pg_proc p WHERE p.pronamespace = 'public'::regnamespace AND p.proname = ? ORDER BY p.oid`

type rpcArg struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type rpcOverload struct {
	setof    bool
	void     bool
	required int
	args     []rpcArg
}

type rpcRepository struct {
	db   *gorm.DB
	meta sync.Map
}

func NewRPCRepository(db *gorm.DB) RPCRepository {
	return &rpcRepository{db: db}
}

// Body JSON dikirim sebagai satu bind lalu dipecah json_to_record dengan tipe dari katalog,
// sama dengan cara PostgREST, supaya konversi tipe dan bentuk hasil identik dengan supabase.rpc.
func (r *rpcRepository) Call(ctx context.Context, fn string, params map[string]json.RawMessage) (json.RawMessage, error) {
	overloads, err := r.overloads(ctx, fn)
	if err != nil {
		return nil, err
	}
	o, ok := pickOverload(overloads, params)
	if !ok {
		return nil, &FunctionNotFoundError{Fn: fn, Keys: slices.Sorted(maps.Keys(params))}
	}

	var vars []any
	if len(params) > 0 {
		body, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		vars = append(vars, string(body))
	}
	var out []byte
	err = r.db.WithContext(ctx).Raw(buildCallSQL(fn, o, params), vars...).Row().Scan(&out)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && out == nil) {
		return json.RawMessage("null"), nil
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *rpcRepository) overloads(ctx context.Context, fn string) ([]rpcOverload, error) {
	if v, ok := r.meta.Load(fn); ok {
		return v.([]rpcOverload), nil
	}
	rows, err := r.db.WithContext(ctx).Raw(rpcMetaSQL, fn).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []rpcOverload
	for rows.Next() {
		var (
			o    rpcOverload
			args []byte
		)
		if err := rows.Scan(&o.setof, &o.void, &o.required, &args); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(args, &o.args); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Fungsi yang belum ada tidak di-cache supaya fungsi baru langsung terpakai tanpa restart.
	if len(list) > 0 {
		r.meta.Store(fn, list)
	}
	return list, nil
}

func pickOverload(overloads []rpcOverload, params map[string]json.RawMessage) (rpcOverload, bool) {
	for _, o := range overloads {
		if acceptsParams(o, params) {
			return o, true
		}
	}
	return rpcOverload{}, false
}

func acceptsParams(o rpcOverload, params map[string]json.RawMessage) bool {
	for key := range params {
		if !slices.ContainsFunc(o.args, func(a rpcArg) bool { return a.Name == key }) {
			return false
		}
	}
	for _, a := range o.args[:min(o.required, len(o.args))] {
		if _, ok := params[a.Name]; !ok {
			return false
		}
	}
	return true
}

func buildCallSQL(fn string, o rpcOverload, params map[string]json.RawMessage) string {
	sel := "to_json(_r)"
	switch {
	case o.setof:
		sel = "coalesce(json_agg(_r), '[]'::json)"
	case o.void:
		sel = "NULL::json"
	}
	call := "public." + quoteIdent(fn)
	if len(params) == 0 {
		return "SELECT " + sel + " FROM " + call + "() AS _r"
	}

	cols := make([]string, 0, len(params))
	named := make([]string, 0, len(params))
	for _, a := range o.args {
		if _, ok := params[a.Name]; !ok {
			continue
		}
		id := quoteIdent(a.Name)
		cols = append(cols, id+" "+a.Type)
		named = append(named, id+" => _a."+id)
	}
	return "SELECT " + sel + " FROM json_to_record(?::json) AS _a(" + strings.Join(cols, ", ") +
		") CROSS JOIN LATERAL " + call + "(" + strings.Join(named, ", ") + ") AS _r"
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
