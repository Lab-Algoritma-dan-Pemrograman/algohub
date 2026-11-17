package jwt

import "testing"

func TestNormalizeRole(t *testing.T) {
	cases := []struct{ raw, want string }{
		{"praktikan", "praktikan"},
		{"mahasiswa", "praktikan"},
		{"mhs", "praktikan"},
		{"penyewa", "penyewa"},
		{"peminjam", "penyewa"},
		{"asisten", "asisten"},
		{"admin", "asisten"},
		{" Admin ", "asisten"},
		{"koordinator", "koordinator"},
		{"kordas", "koordinator"},
		{"KORDA", "koordinator"},
		{"coordinator", "koordinator"},
		{"superadmin", "superadmin"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := NormalizeRole(tc.raw); got != tc.want {
			t.Errorf("NormalizeRole(%q) = %q, ingin %q", tc.raw, got, tc.want)
		}
	}
}
