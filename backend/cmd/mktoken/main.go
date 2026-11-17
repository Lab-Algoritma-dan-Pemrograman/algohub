// Command mktoken membuat JWT uji dari JWT_SECRET di .env, tanpa perlu password.
// Hanya untuk pengujian lokal. Pakai: go run ./cmd/mktoken -uid 1150 -nim 202314020 -role asisten
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
)

func main() {
	uid := flag.Int("uid", 0, "users.id")
	nim := flag.String("nim", "", "NIM")
	role := flag.String("role", "mahasiswa", "role")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tok, err := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpireHours).Generate(*uid, *nim, *role)
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
	fmt.Print(tok)
}
