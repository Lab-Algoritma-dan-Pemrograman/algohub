package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int    `json:"uid"`
	NIM    string `json:"nim"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret      []byte
	expireHours int
}

func NewManager(secret string, expireHours int) *Manager {
	return &Manager{secret: []byte(secret), expireHours: expireHours}
}

// Generate menandatangani token praktikum dengan jti unik (K-1/K-3) supaya
// logout bisa mencabut token spesifik, bukan seluruh sesi user.
func (m *Manager) Generate(userID int, nim, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		NIM:    nim,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expireHours) * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}

// ExpiryTime mengembalikan exp klaim sebagai time.Time; bool kedua false bila
// klaim tidak punya exp (token manual/legacy). Dipakai RevokeToken untuk TTL.
func ExpiryTime(c *Claims) (time.Time, bool) {
	if c == nil || c.ExpiresAt == nil {
		return time.Time{}, false
	}
	return c.ExpiresAt.Time, true
}
