package usecase

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
)

const ssoTTL = 60 * time.Second

var (
	ErrSSOInvalidToken  = errors.New("sso: token tidak valid")
	ErrSSOMissingFields = errors.New("sso: token tanpa nim/nama")

	errSSOUnavailable = errors.New("sso: secret Supabase belum dikonfigurasi")
	errSSOCode        = errors.New("sso: code tidak valid")
)

type ssoPayload struct {
	NIM     any    `json:"nim"`
	Nama    any    `json:"nama"`
	Kelas   any    `json:"kelas"`
	Jurusan any    `json:"jurusan"`
	Email   any    `json:"email"`
	Role    string `json:"role"`
	Exp     int64  `json:"exp"`
	JTI     string `json:"jti"`
}

// Code disegel AES-256-GCM, bukan disimpan di tabel, karena sso_handshakes milik DB siakad dan DDL dilarang (keputusan 8).
// Sekali-pakai hanya dijamin per instance; lintas instance Vercel dibatasi umur 60 detik.
type ssoSealer struct {
	aead cipher.AEAD
	now  func() time.Time
	mu   sync.Mutex
	used map[string]time.Time
}

func newSSOSealer(secret string, now func() time.Time) *ssoSealer {
	s := &ssoSealer{now: now, used: map[string]time.Time{}}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return s
	}
	key, err := hkdf.Key(sha256.New, []byte(secret), nil, "elearning-sso", 32)
	if err != nil {
		return s
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return s
	}
	if aead, err := cipher.NewGCMWithRandomNonce(block); err == nil {
		s.aead = aead
	}
	return s
}

func (s *ssoSealer) seal(p *ssoPayload) (string, error) {
	if s.aead == nil {
		return "", errSSOUnavailable
	}
	p.JTI = rand.Text()
	p.Exp = s.now().Add(ssoTTL).UnixMilli()
	plain, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(s.aead.Seal(nil, nil, plain, nil)), nil
}

func (s *ssoSealer) open(code string) (*ssoPayload, error) {
	if s.aead == nil {
		return nil, errSSOUnavailable
	}
	raw, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return nil, errSSOCode
	}
	plain, err := s.aead.Open(nil, nil, raw, nil)
	if err != nil {
		return nil, errSSOCode
	}
	var p ssoPayload
	dec := json.NewDecoder(bytes.NewReader(plain))
	dec.UseNumber()
	if err := dec.Decode(&p); err != nil || p.JTI == "" {
		return nil, errSSOCode
	}
	now := s.now()
	if now.UnixMilli() > p.Exp {
		return nil, errSSOCode
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for jti, exp := range s.used {
		if now.After(exp) {
			delete(s.used, jti)
		}
	}
	if _, dup := s.used[p.JTI]; dup {
		return nil, errSSOCode
	}
	s.used[p.JTI] = time.UnixMilli(p.Exp)
	return &p, nil
}

// Pengganti api/index.ts: token diverifikasi dan dinormalisasi sekarang, re-sign dilakukan saat code ditukar.
func (uc *AuthUsecase) SSO(token string) (string, error) {
	claims, err := uc.verifyToken(token)
	if err != nil {
		return "", ErrSSOInvalidToken
	}
	id := readIdentity(claims)
	if !truthy(id.nim) || !truthy(id.nama) {
		return "", ErrSSOMissingFields
	}
	return uc.sso.seal(&ssoPayload{NIM: id.nim, Nama: id.nama, Kelas: id.kelas, Jurusan: id.jurusan, Email: id.email, Role: id.role})
}

// Claim re-sign sama dengan api/index.ts (tanpa shift dan uid).
func (uc *AuthUsecase) ReceiveCode(code string) (*dto.ReceiveTokenResponse, error) {
	p, err := uc.sso.open(code)
	if err != nil {
		return nil, Unauthorized("Kode SSO tidak valid atau kedaluwarsa.")
	}
	id := identity{nim: p.NIM, nama: p.Nama, kelas: p.Kelas, jurusan: p.Jurusan, email: p.Email, role: p.Role}
	claims := id.supabaseClaims()
	delete(claims, "shift")
	return uc.receiveResponse("", &id, claims)
}
