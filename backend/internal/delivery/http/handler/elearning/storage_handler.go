package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
)

type StorageHandler struct {
	uc *usecase.StorageUsecase
}

func NewStorageHandler(uc *usecase.StorageUsecase) *StorageHandler {
	return &StorageHandler{uc: uc}
}

// Multipart field "file"; tipe diambil dari Content-Type bagian file seperti File.type di browser.
func (h *StorageHandler) UploadLessonImage(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	f, err := fh.Open()
	if err != nil {
		fail(c, err)
		return
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		fail(c, err)
		return
	}
	img := &dto.ImageFile{Name: fh.Filename, ContentType: fh.Header.Get("Content-Type"), Content: content}
	res, err := h.uc.UploadLessonImage(middleware.ElearningPrincipal(c), img)
	reply(c, "Gambar diunggah", res, err)
}
