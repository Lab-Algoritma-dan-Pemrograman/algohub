package handler

import (
	"net/http"
	"strconv"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	uc *usecase.GameUsecase
}

func NewGameHandler(uc *usecase.GameUsecase) *GameHandler { return &GameHandler{uc: uc} }

// @Summary	Konfigurasi Game
// @Tags		Game
// @Security	bearerAuth
// @Produce	json
// @Success	200	{object}	response.Envelope{data=entity.GameKonfigurasi}
// @Router		/praktikum/game/konfigurasi [get]
func (h *GameHandler) Konfigurasi(c *gin.Context) {
	response.OK(c, http.StatusOK, "Konfigurasi game", h.uc.Konfigurasi())
}

// @Summary		Status Kuota Main
// @Description	Apakah user masih boleh main minggu ini
// @Tags			Game
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope{data=dto.StatusMainResponse}
// @Router			/praktikum/game/status [get]
func (h *GameHandler) StatusMain(c *gin.Context) {
	res, err := h.uc.StatusMain(middleware.UserID(c))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Status main", res)
}

// @Summary		Mulai Bug Hunt
// @Description	Mengambil soal acak TANPA kunci jawaban
// @Tags			Game
// @Security		bearerAuth
// @Produce		json
// @Param			bahasa	query		string	true	"c atau python"
// @Param			jumlah	query		int		false	"jumlah soal"
// @Success		200		{object}	response.Envelope{data=dto.MulaiMainResponse}
// @Router			/praktikum/game/mulai [post]
func (h *GameHandler) MulaiMain(c *gin.Context) {
	jumlah, _ := strconv.Atoi(c.DefaultQuery("jumlah", "5"))
	res, err := h.uc.MulaiMain(middleware.UserID(c), c.Query("bahasa"), jumlah)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Soal siap", res)
}

// @Summary		Kirim Jawaban Bug Hunt
// @Description	Client mengirim JAWABAN, bukan skor. Penilaian dan XP dihitung server.
// @Tags			Game
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.SelesaiMainRequest	true	"Jawaban"
// @Success		200		{object}	response.Envelope{data=dto.SelesaiMainResponse}
// @Router			/praktikum/game/selesai [post]
func (h *GameHandler) SelesaiMain(c *gin.Context) {
	var req dto.SelesaiMainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.uc.SelesaiMain(middleware.UserID(c), req)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Hasil permainan", res)
}

// @Summary	Papan Peringkat
// @Tags		Game
// @Security	bearerAuth
// @Produce	json
// @Param		limit		query		int	false	"jumlah baris"
// @Param		kelas_id	query		int	false	"saring per kelas"
// @Success	200			{object}	response.Envelope
// @Router		/praktikum/game/peringkat [get]
func (h *GameHandler) Peringkat(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	kelasID := queryIntPtr(c, "kelas_id")
	namaKelas := c.Query("kelas")
	baris, err := h.uc.Peringkat(limit, kelasID, namaKelas)
	if err != nil {
		mapError(c, err)
		return
	}
	posisi, err := h.uc.PeringkatSaya(middleware.UserID(c), kelasID, namaKelas)
	if err != nil {
		posisi = 0
	}
	response.OK(c, http.StatusOK, "Papan peringkat", gin.H{
		"peringkat":   baris,
		"posisi_saya": posisi,
	})
}

// @Summary		Heartbeat Sesi Aktif
// @Description	Menandai user sedang online. Identitas diambil dari token.
// @Tags			Game
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.DetakRequest	false	"Aktivitas"
// @Success		200		{object}	response.Envelope
// @Router			/praktikum/game/detak [post]
func (h *GameHandler) Detak(c *gin.Context) {
	var req dto.DetakRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.uc.Detak(middleware.UserID(c), req.Aktivitas); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Detak tercatat", nil)
}

// @Summary		Daftar Sesi Aktif
// @Description	User yang berdetak dalam 2 menit terakhir
// @Tags			Admin - Monitoring
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope
// @Router			/praktikum/admin/monitoring/sesi [get]
func (h *GameHandler) SesiAktif(c *gin.Context) {
	res, err := h.uc.SesiAktif()
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Sesi aktif", res)
}

// @Summary	Daftar Soal Game (dengan kunci jawaban)
// @Tags		Admin - Game
// @Security	bearerAuth
// @Produce	json
// @Param		bahasa	query		string	false	"c atau python"
// @Success	200		{object}	response.Envelope{data=[]entity.GameSoal}
// @Router		/praktikum/admin/game/soal [get]
func (h *GameHandler) AdminListSoal(c *gin.Context) {
	res, err := h.uc.AdminListSoal(c.Query("bahasa"))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Daftar soal game", res)
}

// @Summary	Tambah Soal Game
// @Tags		Admin - Game
// @Security	bearerAuth
// @Accept		json
// @Produce	json
// @Param		request	body		dto.GameSoalRequest	true	"Payload"
// @Success	201		{object}	response.Envelope{data=entity.GameSoal}
// @Router		/praktikum/admin/game/soal [post]
func (h *GameHandler) AdminCreateSoal(c *gin.Context) {
	var req dto.GameSoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.uc.AdminCreateSoal(&req)
	if err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, "Soal dibuat", res)
}

// @Summary	Perbarui Soal Game
// @Tags		Admin - Game
// @Security	bearerAuth
// @Accept		json
// @Produce	json
// @Param		id		path		int					true	"ID Soal"
// @Param		request	body		dto.GameSoalRequest	true	"Payload"
// @Success	200		{object}	response.Envelope{data=entity.GameSoal}
// @Router		/praktikum/admin/game/soal/{id} [put]
func (h *GameHandler) AdminUpdateSoal(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.GameSoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.uc.AdminUpdateSoal(id, &req)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Soal diperbarui", res)
}

// @Summary	Hapus Soal Game
// @Tags		Admin - Game
// @Security	bearerAuth
// @Produce	json
// @Param		id	path		int	true	"ID Soal"
// @Success	200	{object}	response.Envelope
// @Router		/praktikum/admin/game/soal/{id} [delete]
func (h *GameHandler) AdminDeleteSoal(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.uc.AdminDeleteSoal(id); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Soal dihapus", nil)
}

// @Summary	Ubah Konfigurasi Game
// @Tags		Admin - Game
// @Security	bearerAuth
// @Accept		json
// @Produce	json
// @Param		request	body		dto.GameKonfigurasiRequest	true	"Payload"
// @Success	200		{object}	response.Envelope{data=entity.GameKonfigurasi}
// @Router		/praktikum/admin/game/konfigurasi [put]
func (h *GameHandler) AdminSimpanKonfigurasi(c *gin.Context) {
	var req dto.GameKonfigurasiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.uc.SimpanKonfigurasi(req)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Konfigurasi disimpan", res)
}
