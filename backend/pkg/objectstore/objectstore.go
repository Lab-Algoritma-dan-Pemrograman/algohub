package objectstore

import (
	"errors"
	"fmt"
	"time"
)

// Client adalah operasi penyimpanan berkas yang dipakai aplikasi. Implementasi
// saat ini adalah Supabase Storage dan Vercel Blob.
type Client interface {
	Enabled() bool
	Upload(path string, content []byte, contentType string, upsert bool) (string, error)
	SignedUploadURL(path string, ttl time.Duration) (string, error)
	SignedDownloadURL(path string, ttl time.Duration, downloadName string) (string, error)
	Delete(path string) error
}

// Failover mencoba setiap penyimpanan yang aktif sesuai urutan. Vercel Blob
// ditempatkan pertama di produksi selama gateway Supabase sedang menolak 402;
// Supabase tetap dipertahankan agar konfigurasi lama tidak langsung putus.
type Failover struct {
	clients []Client
}

func NewFailover(clients ...Client) *Failover {
	return &Failover{clients: clients}
}

func (f *Failover) Enabled() bool {
	for _, client := range f.clients {
		if client != nil && client.Enabled() {
			return true
		}
	}
	return false
}

func (f *Failover) Upload(path string, content []byte, contentType string, upsert bool) (string, error) {
	return firstResult(f.clients, func(client Client) (string, error) {
		return client.Upload(path, content, contentType, upsert)
	})
}

func (f *Failover) SignedUploadURL(path string, ttl time.Duration) (string, error) {
	return firstResult(f.clients, func(client Client) (string, error) {
		return client.SignedUploadURL(path, ttl)
	})
}

func (f *Failover) SignedDownloadURL(path string, ttl time.Duration, downloadName string) (string, error) {
	return firstResult(f.clients, func(client Client) (string, error) {
		return client.SignedDownloadURL(path, ttl, downloadName)
	})
}

func (f *Failover) Delete(path string) error {
	var errs []error
	for _, client := range f.clients {
		if client == nil || !client.Enabled() {
			continue
		}
		if err := client.Delete(path); err == nil {
			return nil
		} else {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return fmt.Errorf("penyimpanan berkas belum dikonfigurasi")
	}
	return errors.Join(errs...)
}

func firstResult(clients []Client, call func(Client) (string, error)) (string, error) {
	var errs []error
	for _, client := range clients {
		if client == nil || !client.Enabled() {
			continue
		}
		result, err := call(client)
		if err == nil {
			return result, nil
		}
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return "", fmt.Errorf("penyimpanan berkas belum dikonfigurasi")
	}
	return "", errors.Join(errs...)
}
