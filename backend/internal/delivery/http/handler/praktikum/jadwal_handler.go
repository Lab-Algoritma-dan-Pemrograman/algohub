package handler

import "github.com/gin-gonic/gin"

// @Summary		Daftar Jadwal Praktikum
// @Description	Mengambil daftar jadwal praktikum laboratorium
// @Tags			Info, Admin - Jadwal
// @Produce		json
// @Success		200	{object}	response.Envelope{data=[]entity.Jadwal}
// @Router			/praktikum/info/jadwal [get]
// @Router			/praktikum/admin/jadwal [get]
func (h *JadwalHandler) List(c *gin.Context) { h.handleList(c) }

// @Summary		Buat Jadwal Baru
// @Description	Menambahkan jadwal praktikum baru (Admin)
// @Tags			Admin - Jadwal
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.JadwalRequest	true	"Payload Jadwal"
// @Success		201		{object}	response.Envelope{data=entity.Jadwal}
// @Router			/praktikum/admin/jadwal [post]
func (h *JadwalHandler) Create(c *gin.Context) { h.handleCreate(c) }

// @Summary		Perbarui Jadwal
// @Description	Mengubah data jadwal praktikum yang ada (Admin)
// @Tags			Admin - Jadwal
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			id		path		int					true	"ID Jadwal"
// @Param			request	body		dto.JadwalRequest	true	"Payload Jadwal"
// @Success		200		{object}	response.Envelope{data=entity.Jadwal}
// @Router			/praktikum/admin/jadwal/{id} [put]
func (h *JadwalHandler) Update(c *gin.Context) { h.handleUpdate(c) }

// @Summary		Hapus Jadwal
// @Description	Menghapus jadwal praktikum (Admin)
// @Tags			Admin - Jadwal
// @Security		bearerAuth
// @Produce		json
// @Param			id	path		int	true	"ID Jadwal"
// @Success		200	{object}	response.Envelope
// @Router			/praktikum/admin/jadwal/{id} [delete]
func (h *JadwalHandler) Delete(c *gin.Context) { h.handleDelete(c) }
