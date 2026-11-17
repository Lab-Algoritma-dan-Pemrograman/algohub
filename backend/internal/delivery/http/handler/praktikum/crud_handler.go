package handler

import (
	"net/http"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type crudPesan struct{ list, create, update, del string }

type crudHandler[T, R, L any] struct {
	list   func() (L, error)
	create func(*T) (R, error)
	update func(int, *T) (R, error)
	del    func(int) error
	pesan  crudPesan
}

func newCrud[T, R, L any](
	list func() (L, error), create func(*T) (R, error), update func(int, *T) (R, error), del func(int) error, pesan crudPesan,
) crudHandler[T, R, L] {
	return crudHandler[T, R, L]{list: list, create: create, update: update, del: del, pesan: pesan}
}

func (h *crudHandler[T, R, L]) handleList(c *gin.Context) {
	res, err := h.list()
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, h.pesan.list, res)
}

func (h *crudHandler[T, R, L]) handleCreate(c *gin.Context) {
	if _, res, ok := bindCreate(c, h.create); ok {
		response.Created(c, h.pesan.create, res)
	}
}

func (h *crudHandler[T, R, L]) handleUpdate(c *gin.Context) {
	if res, ok := bindUpdate(c, "id", h.update); ok {
		response.OK(c, http.StatusOK, h.pesan.update, res)
	}
}

func (h *crudHandler[T, R, L]) handleDelete(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.del(id); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, h.pesan.del, nil)
}

type (
	JadwalHandler struct {
		crudHandler[dto.JadwalRequest, *entity.Jadwal, []entity.Jadwal]
	}
	KelasHandler struct {
		crudHandler[dto.KelasRequest, *entity.Kelas, []entity.Kelas]
	}
	PedomanHandler struct {
		crudHandler[dto.PedomanRequest, *entity.PedomanLaporan, []entity.PedomanLaporan]
	}
)

func NewJadwalHandler(uc *usecase.JadwalUsecase) *JadwalHandler {
	return &JadwalHandler{newCrud(uc.List, uc.Create, uc.Update, uc.Delete,
		crudPesan{"Daftar jadwal", "Jadwal dibuat", "Jadwal diperbarui", "Jadwal dihapus"})}
}

func NewKelasHandler(uc *usecase.KelasUsecase) *KelasHandler {
	return &KelasHandler{newCrud(uc.List, uc.Create, uc.Update, uc.Delete,
		crudPesan{"Daftar kelas", "Kelas dibuat", "Kelas diperbarui", "Kelas dihapus"})}
}

func NewPedomanHandler(uc *usecase.PedomanUsecase) *PedomanHandler {
	return &PedomanHandler{newCrud(uc.List, uc.Create, uc.Update, uc.Delete,
		crudPesan{"Daftar pedoman", "Pedoman ditambahkan", "Pedoman diperbarui", "Pedoman dihapus"})}
}
