package usecase

import (
	"encoding/json"
	"testing"
)

func TestStripCredentials(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"objek", `{"id": 1, "password": "h", "nama": "<A&B>"}`, `{"id":1,"nama":"<A&B>"}`},
		{"array", `[{"id":1,"fb_password_salt":"s"}, 5, null, {"id":2}]`, `[{"id":1},5,null,{"id":2}]`},
		{"angka", `42`, `42`},
		{"null", `null`, `null`},
		{"string", `"password"`, `"password"`},
		{"tanpa kredensial tetap byte identik", "{\"b\": 1,  \"a\": [1, 2]}", "{\"b\": 1,  \"a\": [1, 2]}"},
		{"array tanpa kredensial", "[{\"b\": 1}, \n {\"a\": 2}]", "[{\"b\": 1}, \n {\"a\": 2}]"},
		{"bersarang tidak dibersihkan", `{"user":{"password":"h"}}`, `{"user":{"password":"h"}}`},
		{"angka besar utuh", `{"id":9007199254740993,"password":"h"}`, `{"id":9007199254740993}`},
	}
	for _, tc := range cases {
		got, err := stripCredentials(json.RawMessage(tc.in))
		if err != nil || string(got) != tc.want {
			t.Errorf("%s: %s, %v; ingin %s", tc.name, got, err, tc.want)
		}
	}
}

func TestTruthy(t *testing.T) {
	cases := []struct {
		v    any
		want bool
	}{
		{nil, false}, {false, false}, {true, true}, {"", false}, {"x", true},
		{json.Number("0"), false}, {json.Number("0.0"), false}, {json.Number("3"), true},
		{map[string]any{}, true}, {[]any{}, true},
	}
	for _, tc := range cases {
		if got := truthy(tc.v); got != tc.want {
			t.Errorf("truthy(%#v) = %v, ingin %v", tc.v, got, tc.want)
		}
	}
}
