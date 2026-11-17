package usecase_test

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestUploadLessonImagePath(t *testing.T) {
	cases := map[string]string{"foto.jpeg": "jpeg", "a.b.webp": "webp", "tanpa-titik": "tanpa-titik", "akhir.": "jpg", "": "jpg"}
	for name, ext := range cases {
		f := newLearn(t)
		_, err := f.storageUC.UploadLessonImage(who(usecase.RoleAsisten), &dto.ImageFile{Name: name, ContentType: "image/jpeg"})
		require.NoError(t, err)
		want := regexp.MustCompile(fmt.Sprintf(`^lessons/%d-[0-9a-z]{5}\.%s$`, learnNow.UnixMilli(), regexp.QuoteMeta(ext)))
		assert.Regexp(t, want, f.uploader.path, name)
	}
}

func TestUploadLessonImageBucketMissing(t *testing.T) {
	f := newLearn(t)
	f.uploader.err = errors.New(`upload gagal (404): {"statusCode":"404","error":"Bucket not found","message":"Bucket not found"}`)
	_, err := f.storageUC.UploadLessonImage(who(usecase.RoleAdmin), pngFile())
	assertUCError(t, err, http.StatusInternalServerError, `Bucket "lesson-images" tidak ditemukan di Supabase. Silakan buat bucket tersebut di dashboard Supabase Anda dan atur agar publik.`)

	f.uploader.err = errors.New("upload gagal (500): lain")
	_, err = f.storageUC.UploadLessonImage(who(usecase.RoleAdmin), pngFile())
	assert.EqualError(t, err, "upload gagal (500): lain")
}
