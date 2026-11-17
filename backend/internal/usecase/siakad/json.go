package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// JSON hanya di-encode ulang kalau ada key yang benar-benar dibuang, supaya urutan kolom
// dari Postgres tetap utuh di kasus umum.
func omitKeys(raw json.RawMessage, keys ...string) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, nil
	}
	switch trimmed[0] {
	case '{':
		obj, changed, err := omitObjectKeys(trimmed, keys)
		if err != nil || !changed {
			return raw, err
		}
		return encodeJSON(obj)
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, err
		}
		changed := false
		for i, item := range items {
			item = bytes.TrimSpace(item)
			if len(item) == 0 || item[0] != '{' {
				continue
			}
			obj, ok, err := omitObjectKeys(item, keys)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if items[i], err = encodeJSON(obj); err != nil {
				return nil, err
			}
			changed = true
		}
		if !changed {
			return raw, nil
		}
		return encodeJSON(items)
	}
	return raw, nil
}

func omitObjectKeys(raw json.RawMessage, keys []string) (obj map[string]json.RawMessage, changed bool, err error) {
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, false, err
	}
	for _, k := range keys {
		if _, ok := obj[k]; ok {
			delete(obj, k)
			changed = true
		}
	}
	return obj, changed, nil
}

// Tanpa escape HTML supaya sama dengan JSON.stringify di kode lama.
func encodeJSON(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// C-05: field kredensial dibuang sebelum dikirim ke klien walau RPC-nya lupa memfilter kolom.
// Server tetap menerima objek utuh dari DB.
func stripCredentials(raw json.RawMessage) (json.RawMessage, error) {
	return omitKeys(raw, "password", "password_hash", "fb_password_hash", "fb_password_salt")
}

func decodeObject(raw json.RawMessage) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var obj map[string]any
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("bukan objek JSON: %s", raw)
	}
	return obj, nil
}

// Aturan truthy JavaScript, karena kode lama memakai `if (x)` pada nilai dari DB dan token.
func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case json.Number:
		f, err := x.Float64()
		return err != nil || (f != 0 && !math.IsNaN(f))
	case float64:
		return x != 0 && !math.IsNaN(x)
	}
	return true
}

// Meniru Number(x) JavaScript untuk id dari body JSON: hanya bilangan bulat yang dipakai sebagai angka.
func integerID(v any) (int64, bool) {
	var f float64
	switch x := v.(type) {
	case float64:
		f = x
	case bool:
		if x {
			f = 1
		}
	case string:
		s := strings.TrimSpace(x)
		if s != "" {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, false
			}
			f = n
		}
	default:
		return 0, false
	}
	if f != math.Trunc(f) || math.Abs(f) > 1<<53 {
		return 0, false
	}
	return int64(f), true
}

func claimFloat(v any) (float64, error) {
	if n, ok := v.(json.Number); ok {
		return n.Float64()
	}
	return 0, fmt.Errorf("klaim bukan angka: %v", v)
}

// Nilai klaim dikirim ke parameter text/bigint dalam bentuk teks seperti PostgREST menerimanya.
func claimText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	}
	return fmt.Sprint(v)
}
