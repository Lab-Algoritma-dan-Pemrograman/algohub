package usecase

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Pengganti undefined JavaScript: properti yang tidak ada berbeda dari null pada typeof dan template string api/*.ts.
type jsUndefined struct{}

var undefined any = jsUndefined{}

// Objek JSON dengan urutan kunci terjaga, karena Object.entries di api/grade.ts menentukan urutan kriteria di prompt.
type jsObject struct {
	keys []string
	vals map[string]any
}

func (o *jsObject) get(key string) any {
	if v, ok := o.vals[key]; ok {
		return v
	}
	return undefined
}

// Urutan Object.entries: kunci indeks array (bilangan bulat kanonik) naik lebih dulu, lalu sisanya sesuai urutan sisip.
func (o *jsObject) entries() []string {
	var idx, rest []string
	for _, k := range o.keys {
		if isArrayIndex(k) {
			idx = append(idx, k)
		} else {
			rest = append(rest, k)
		}
	}
	slices.SortFunc(idx, func(a, b string) int {
		x, _ := strconv.ParseUint(a, 10, 32)
		y, _ := strconv.ParseUint(b, 10, 32)
		return cmp.Compare(x, y)
	})
	return append(idx, rest...)
}

func isArrayIndex(k string) bool {
	n, err := strconv.ParseUint(k, 10, 32)
	return err == nil && n < math.MaxUint32 && strconv.FormatUint(n, 10) == k
}

var errJSONTrailing = errors.New("data JSON berlebih setelah nilai")

// Setara JSON.parse: objek menjadi *jsObject, angka json.Number; kunci ganda memakai nilai terakhir di posisi pertama.
func parseJS(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := decodeJS(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errJSONTrailing
	}
	return v, nil
}

func decodeJS(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	if delim == '[' {
		arr := []any{}
		for dec.More() {
			v, err := decodeJS(dec)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		_, err := dec.Token()
		return arr, err
	}
	obj := &jsObject{vals: map[string]any{}}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := kt.(string)
		v, err := decodeJS(dec)
		if err != nil {
			return nil, err
		}
		if _, dup := obj.vals[key]; !dup {
			obj.keys = append(obj.keys, key)
		}
		obj.vals[key] = v
	}
	_, err = dec.Token()
	return obj, err
}

func nullable(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

// Akses properti x.key / x?.[key]; selain objek hasilnya undefined.
func prop(v any, key string) any {
	if o, ok := v.(*jsObject); ok {
		return o.get(key)
	}
	return undefined
}

func jsTruthy(v any) bool {
	if _, ok := v.(jsUndefined); ok {
		return false
	}
	if f, ok := v.(float64); ok {
		return f != 0 && !math.IsNaN(f)
	}
	return truthy(v)
}

// Operator || JavaScript, dipakai untuk nilai default seperti `x || 'N/A'` di api/grade.ts.
func or(a, b any) any {
	if jsTruthy(a) {
		return a
	}
	return b
}

// Konversi String(x) / template literal JavaScript.
func jsToString(v any) string {
	switch x := v.(type) {
	case jsUndefined:
		return "undefined"
	case nil:
		return "null"
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case json.Number:
		return jsNumberString(jsNumber(x))
	case float64:
		return jsNumberString(x)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			if e != nil {
				parts[i] = jsToString(e)
			}
		}
		return strings.Join(parts, ",")
	}
	return "[object Object]"
}

func jsNumberString(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0:
		return "0"
	}
	if a := math.Abs(f); a >= 1e21 || a < 1e-6 {
		s := strconv.FormatFloat(f, 'e', -1, 64)
		mant, exp, _ := strings.Cut(s, "e")
		sign, digits := exp[:1], strings.TrimLeft(exp[1:], "0")
		return mant + "e" + sign + digits
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

var jsDecimal = regexp.MustCompile(`^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$`)

// Konversi Number(x) JavaScript.
func jsNumber(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case bool:
		if x {
			return 1
		}
		return 0
	case float64:
		return x
	case json.Number:
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			return math.NaN()
		}
		return f
	case string:
		return stringToNumber(x)
	case []any:
		switch len(x) {
		case 0:
			return 0
		case 1:
			return stringToNumber(jsToString(x))
		}
	}
	return math.NaN()
}

func stringToNumber(s string) float64 {
	s = strings.TrimSpace(s)
	switch s {
	case "":
		return 0
	case "Infinity", "+Infinity":
		return math.Inf(1)
	case "-Infinity":
		return math.Inf(-1)
	}
	if len(s) > 2 && s[0] == '0' {
		base := map[byte]int{'x': 16, 'X': 16, 'o': 8, 'O': 8, 'b': 2, 'B': 2}[s[1]]
		if base != 0 {
			n, err := strconv.ParseUint(s[2:], base, 64)
			if err != nil {
				return math.NaN()
			}
			return float64(n)
		}
	}
	if !jsDecimal.MatchString(s) {
		return math.NaN()
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return math.NaN()
	}
	return f
}
