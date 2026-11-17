package handler

import "github.com/gin-gonic/gin"

// @Summary		Daftar Pedoman Laporan
// @Description	Mengambil daftar pedoman laporan (Admin dan Publik)
// @Tags			Info, Admin - Pedoman
// @Produce		json
// @Success		200	{object}	response.Envelope{data=[]entity.PedomanLaporan}
// @Router			/praktikum/info/laporan [get]
// @Router			/praktikum/admin/pedoman [get]
func (h *PedomanHandler) List(c *gin.Context) { h.handleList(c) }

// @Summary		Tambah Pedoman
// @Description	Menambahkan pedoman laporan baru
// @Tags			Admin - Pedoman
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.PedomanRequest	true	"Payload Pedoman"
// @Success		201		{object}	response.Envelope{data=entity.PedomanLaporan}
// @Router			/praktikum/admin/pedoman [post]
func (h *PedomanHandler) Create(c *gin.Context) { h.handleCreate(c) }

// @Summary		Perbarui Pedoman
// @Description	Mengubah pedoman laporan
// @Tags			Admin - Pedoman
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			id		path		int					true	"ID Pedoman"
// @Param			request	body		dto.PedomanRequest	true	"Payload Pedoman"
// @Success		200		{object}	response.Envelope{data=entity.PedomanLaporan}
// @Router			/praktikum/admin/pedoman/{id} [put]
func (h *PedomanHandler) Update(c *gin.Context) { h.handleUpdate(c) }

// @Summary		Hapus Pedoman
// @Description	Menghapus pedoman laporan
// @Tags			Admin - Pedoman
// @Security		bearerAuth
// @Produce		json
// @Param			id	path		int	true	"ID Pedoman"
// @Success		200	{object}	response.Envelope
// @Router			/praktikum/admin/pedoman/{id} [delete]
func (h *PedomanHandler) Delete(c *gin.Context) { h.handleDelete(c) }
