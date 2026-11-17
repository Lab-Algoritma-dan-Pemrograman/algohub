package usecase

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/supabasejwt"
)

// setExpirationTime('1d') di api/*.ts lama.
const tokenTTL = 24 * time.Hour

var errNoVerifier = errors.New("konfigurasi verifikasi token belum tersedia")

type PraktikumLoginResult struct {
	Token     string
	NIM       string
	Nama      string
	NamaKelas string
	Shift     *int
	Role      string
}

// Login langsung tidak lagi memproksi HTTP ke api.algohub.web.id; perakitan ke usecase praktikum ada di internal/app (keputusan 9).
type PraktikumLogin interface {
	Login(ctx context.Context, identifier, password, ip, userAgent string) (*PraktikumLoginResult, error)
}

type AuthUsecase struct {
	cfg     config.ElearningConfig
	login   PraktikumLogin
	sso     *ssoSealer
	keyOnce sync.Once
	key     []byte
}

func NewAuthUsecase(cfg *config.ElearningConfig, login PraktikumLogin, now func() time.Time) *AuthUsecase {
	return &AuthUsecase{cfg: *cfg, login: login, sso: newSSOSealer(cfg.SupabaseJWTSecret, now)}
}

// Deteksi encoding cukup sekali per instance, sama dengan cache _secretDetected di api/auth.ts.
func (uc *AuthUsecase) supabaseKey() []byte {
	uc.keyOnce.Do(func() {
		uc.key = supabasejwt.DetectSecret(uc.cfg.SupabaseJWTSecret, uc.cfg.SupabaseAnonKey)
	})
	return uc.key
}

// Urutan sama dengan verifyToken di api/auth.ts: secret Supabase (semua HS), lalu JWT_SECRET dan SIAKAD_JWT_SECRET (HS256 saja).
// Error yang dikembalikan berasal dari secret legacy terakhir, seperti lastError di kode lama.
func (uc *AuthUsecase) verifyToken(token string) (jwt.MapClaims, error) {
	if key := uc.supabaseKey(); key != nil {
		if claims, err := supabasejwt.VerifyHMAC(token, key); err == nil {
			return claims, nil
		}
	}
	lastErr := errNoVerifier
	for _, secret := range []string{uc.cfg.LegacyJWTSecret, uc.cfg.SiakadJWTSecret} {
		if secret == "" {
			continue
		}
		claims, err := supabasejwt.VerifyHMAC(token, []byte(secret), "HS256")
		if err == nil {
			return claims, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// Teks meniru pesan jose supaya "Invalid token: ..." dan "Token tidak valid: ..." sama dengan api/*.ts lama.
func joseMessage(err error) string {
	switch {
	case errors.Is(err, errNoVerifier):
		return "Konfigurasi verifikasi token belum tersedia."
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "Invalid Compact JWS"
	case errors.Is(err, jwt.ErrTokenExpired):
		return `"exp" claim timestamp check failed`
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return `"nbf" claim timestamp check failed`
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "signature verification failed"
	}
	return err.Error()
}

// Tanpa secret Supabase token asli dikembalikan apa adanya, sama dengan api/*.ts lama.
func (uc *AuthUsecase) resign(original string, claims jwt.MapClaims) (string, error) {
	key := uc.supabaseKey()
	if key == nil {
		return original, nil
	}
	return supabasejwt.SignHS256(claims, key, tokenTTL)
}

type identity struct {
	uid, nim, nama, kelas, jurusan, email any
	shift                                 *string
	role                                  string
}

func readIdentity(c jwt.MapClaims) identity {
	return identity{
		uid:     pick(c, nil, "uid"),
		nim:     pick(c, nil, "nim", "username", "sub"),
		nama:    pick(c, nil, "nama", "full_name", "name"),
		kelas:   pick(c, "", "kelas", "class_code"),
		jurusan: pick(c, nil, "jurusan", "major"),
		email:   pick(c, nil, "email"),
		shift:   shiftOf(c["shift"]),
		role:    NormalizeTokenRole(jsString(pick(c, "praktikan", "user_role", "role"))),
	}
}

func (id *identity) supabaseClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"nim":       id.nim,
		"nama":      id.nama,
		"kelas":     id.kelas,
		"shift":     id.shift,
		"jurusan":   id.jurusan,
		"email":     id.email,
		"role":      "authenticated",
		"user_role": id.role,
		"iss":       "supabase",
		"sub":       id.nim,
		"aud":       "authenticated",
	}
}

func (id *identity) payload() dto.AuthPayload {
	return dto.AuthPayload{NIM: id.nim, Nama: id.nama, Kelas: id.kelas, Shift: id.shift, Jurusan: id.jurusan, Email: id.email, Role: id.role}
}

// Meniru operator || JavaScript: nilai falsy (null, "", 0, false) dilewati.
func pick(c map[string]any, fallback any, keys ...string) any {
	for _, k := range keys {
		if truthy(c[k]) {
			return c[k]
		}
	}
	return fallback
}

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
		return err != nil || f != 0
	case float64:
		return x != 0
	}
	return true
}

func jsString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	}
	return fmt.Sprint(v)
}

// shift ?? null, lalu String(shift) kosong dianggap null (api/receive-token.ts).
func shiftOf(v any) *string {
	if v == nil {
		return nil
	}
	s := jsString(v)
	if s == "" {
		return nil
	}
	return &s
}

func optString(v any) *string {
	if v == nil {
		return nil
	}
	s := jsString(v)
	return &s
}

// Urutan kunci claim tidak terjaga di map Go, jadi receivedKeys diurutkan alfabetis.
func missingFields(message string, c jwt.MapClaims) *Error {
	e := BadRequest(message)
	e.Detail = map[string]any{"receivedKeys": slices.Sorted(maps.Keys(c))}
	return e
}

func (uc *AuthUsecase) Authenticate(token string) (*dto.Principal, error) {
	claims, err := uc.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}
	id := readIdentity(claims)
	p := &dto.Principal{
		NIM:       jsString(id.nim),
		AuthNIM:   authNIM(claims),
		UID:       jsString(id.uid),
		Nama:      jsString(id.nama),
		Kelas:     jsString(id.kelas),
		Shift:     id.shift,
		Jurusan:   optString(id.jurusan),
		Email:     optString(id.email),
		Role:      id.role,
		RawClaims: claims,
	}
	if p.NIM == "" && p.AuthNIM == "" {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}
	if p.Nama == "" {
		p.Nama = p.NIM
	}
	return p, nil
}

// public.auth_nim(): COALESCE(claims->>'nim', claims->>'sub', claims->>'username'); string kosong tidak dilewati.
func authNIM(c jwt.MapClaims) string {
	for _, k := range []string{"nim", "sub", "username"} {
		if v, ok := c[k]; ok && v != nil {
			return jsString(v)
		}
	}
	return ""
}

func (uc *AuthUsecase) Verify(token string) (*dto.VerifyResponse, error) {
	if token == "" {
		return nil, BadRequest("Token is required")
	}
	claims, err := uc.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Invalid token: " + joseMessage(err))
	}
	id := readIdentity(claims)
	// Token Go praktikum (uid, nim, role) tidak membawa nama; nama hanya untuk tampilan jadi diturunkan dari nim.
	if !truthy(id.nama) && truthy(id.nim) {
		id.nama = jsString(id.nim)
	}
	if !truthy(id.nim) || !truthy(id.nama) {
		return nil, missingFields("Token missing required fields (nim/username, nama/full_name)", claims)
	}
	c := id.supabaseClaims()
	c["uid"] = id.uid
	signed, err := uc.resign(token, c)
	if err != nil {
		return nil, Unauthorized("Invalid token: " + err.Error())
	}
	return &dto.VerifyResponse{Payload: id.payload(), Token: signed}, nil
}

func (uc *AuthUsecase) ReceiveToken(token string) (*dto.ReceiveTokenResponse, error) {
	if token == "" {
		return nil, BadRequest(`Field "token" wajib disertakan di request body.`)
	}
	claims, err := uc.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Token tidak valid: " + joseMessage(err))
	}
	id := readIdentity(claims)
	if !truthy(id.nim) || !truthy(id.nama) {
		return nil, missingFields("Token tidak memiliki field yang diperlukan (nim/username, nama/full_name).", claims)
	}
	return uc.receiveResponse(token, &id, id.supabaseClaims())
}

func (uc *AuthUsecase) receiveResponse(original string, id *identity, claims jwt.MapClaims) (*dto.ReceiveTokenResponse, error) {
	signed, err := uc.resign(original, claims)
	if err != nil {
		return nil, Unauthorized("Token tidak valid: " + err.Error())
	}
	return &dto.ReceiveTokenResponse{
		Success:     true,
		Token:       signed,
		RedirectURL: "/?token=" + url.QueryEscape(signed),
		Payload:     id.payload(),
	}, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" || req.Password == "" {
		return nil, BadRequest("NIM/email dan password wajib diisi.")
	}
	res, err := uc.login.Login(ctx, identifier, req.Password, ip, userAgent)
	if err != nil {
		// MEDIUM-01: pesan generik — jangan bedakan "akun tidak ditemukan" vs
		// "password salah" (mencegah enumerasi NIM/email yang valid).
		return nil, Unauthorized("NIM/email atau password salah.")
	}
	nim := cmp.Or(res.NIM, identifier)
	nama := cmp.Or(res.Nama, nim)
	var shift *string
	if res.Shift != nil {
		s := strconv.Itoa(*res.Shift)
		shift = &s
	}
	id := identity{nim: nim, nama: nama, kelas: res.NamaKelas, shift: shift, role: NormalizeTokenRole(cmp.Or(res.Role, "mahasiswa"))}
	signed, err := uc.resign(res.Token, id.supabaseClaims())
	if err != nil {
		return nil, newError(http.StatusInternalServerError, "Terjadi kesalahan server. Coba lagi.")
	}
	var namaKelas *string
	if res.NamaKelas != "" {
		namaKelas = &res.NamaKelas
	}
	return &dto.LoginResponse{
		Payload:      id.payload(),
		Token:        signed,
		BackendToken: res.Token,
		User:         dto.LoginUser{NIM: nim, Nama: nama, Kelas: res.NamaKelas, Role: id.role, NamaKelas: namaKelas, Shift: res.Shift},
	}, nil
}

// Token Realtime dibuat dari principal yang sudah terverifikasi, dengan claim yang sama seperti receive-token.
func (uc *AuthUsecase) SupabaseToken(token string, p *dto.Principal) (*dto.SupabaseTokenResponse, error) {
	if uc.supabaseKey() == nil {
		return &dto.SupabaseTokenResponse{Token: token}, nil
	}
	id := identity{nim: p.NIM, nama: p.Nama, kelas: p.Kelas, shift: p.Shift, role: p.Role}
	if p.Jurusan != nil {
		id.jurusan = *p.Jurusan
	}
	if p.Email != nil {
		id.email = *p.Email
	}
	signed, err := uc.resign(token, id.supabaseClaims())
	if err != nil {
		return nil, err
	}
	return &dto.SupabaseTokenResponse{Token: signed, ExpiresIn: int(tokenTTL / time.Second)}, nil
}
