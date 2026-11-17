package usecase

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"slices"
	"strings"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
)

const (
	msgImageFormat    = "Format file tidak didukung. Harap upload gambar (JPG, PNG, GIF, WEBP)."
	msgBucketNotFound = `Bucket "lesson-images" tidak ditemukan di Supabase. Silakan buat bucket tersebut di dashboard Supabase Anda dan atur agar publik.`
	// Teks Storage API tanpa nama tabel (elearning-rls.md § Status tidak pasti 10).
	msgStorageRLS = "new row violates row-level security policy"
)

var imageTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

type ImageUploader interface {
	Upload(path string, content []byte, contentType string, upsert bool) (string, error)
}

type StorageUsecase struct {
	uploader ImageUploader
	now      func() time.Time
}

func NewStorageUsecase(uploader ImageUploader, now func() time.Time) *StorageUsecase {
	return &StorageUsecase{uploader: uploader, now: now}
}

// Disalin dari storageService.uploadImage. Upload memakai service role key yang melewati policy storage.objects,
// jadi syarat staf (R-storage_objects-INSERT-1) dicek di sini sesudah validasi tipe file seperti urutan lama.
func (uc *StorageUsecase) UploadLessonImage(p *dto.Principal, f *dto.ImageFile) (*dto.UploadResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	if !slices.Contains(imageTypes, f.ContentType) {
		return nil, BadRequest(msgImageFormat)
	}
	if !IsStaff(p.Role) {
		return nil, Forbidden(msgStorageRLS)
	}
	url, err := uc.uploader.Upload(lessonImagePath(f.Name, uc.now()), f.Content, f.ContentType, false)
	if err != nil {
		if strings.Contains(err.Error(), "Bucket not found") {
			return nil, newError(http.StatusInternalServerError, msgBucketNotFound)
		}
		return nil, err
	}
	return &dto.UploadResponse{PublicURL: url}, nil
}

// `${Date.now()}-${Math.random().toString(36).substring(2, 7)}.${file.name.split('.').pop() || 'jpg'}`.
func lessonImagePath(name string, now time.Time) string {
	ext := name[strings.LastIndex(name, ".")+1:]
	if ext == "" {
		ext = "jpg"
	}
	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	suffix := make([]byte, 5)
	for i := range suffix {
		suffix[i] = base36[rand.IntN(len(base36))]
	}
	return fmt.Sprintf("lessons/%d-%s.%s", now.UnixMilli(), suffix, ext)
}
