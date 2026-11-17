package handler

import "github.com/gin-gonic/gin"

// @Summary		Daftar Kelas
// @Description	Mengambil daftar kelas yang tersedia
// @Tags			Admin - Kelas
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope{data=[]entity.Kelas}
// @Router			/praktikum/admin/kelas [get]
func (h *KelasHandler) List(c *gin.Context) { h.handleList(c) }

// @Summary		Tambah Kelas
// @Description	Menambahkan kelas baru
// @Tags			Admin - Kelas
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.KelasRequest	true	"Payload"
// @Success		201		{object}	response.Envelope{data=entity.Kelas}
// @Router			/praktikum/admin/kelas [post]
func (h *KelasHandler) Create(c *gin.Context) { h.handleCreate(c) }

// @Summary		Perbarui Kelas
// @Description	Mengubah data kelas
// @Tags			Admin - Kelas
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			id		path		int					true	"ID Kelas"
// @Param			request	body		dto.KelasRequest	true	"Payload"
// @Success		200		{object}	response.Envelope{data=entity.Kelas}
// @Router			/praktikum/admin/kelas/{id} [put]
func (h *KelasHandler) Update(c *gin.Context) { h.handleUpdate(c) }

// @Summary		Hapus Kelas
// @Description	Menghapus kelas
// @Tags			Admin - Kelas
// @Security		bearerAuth
// @Produce		json
// @Param			id	path		int	true	"ID Kelas"
// @Success		200	{object}	response.Envelope
// @Router			/praktikum/admin/kelas/{id} [delete]
func (h *KelasHandler) Delete(c *gin.Context) { h.handleDelete(c) }
