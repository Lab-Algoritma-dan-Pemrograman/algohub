package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/app"
)

// Vercel mengisi PORT. Tanpa PORT (laptop), listen di 127.0.0.1 supaya Windows Firewall tidak meminta izin.
const localAddr = "127.0.0.1:8080"

func main() {
	handler, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	addr := localAddr
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("algohub-api listen di %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
