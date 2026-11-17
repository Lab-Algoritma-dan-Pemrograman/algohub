package handler

import (
	"net/http"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth     *usecase.AuthUsecase
	profile  *usecase.ProfileUsecase
	auditLog *usecase.AuditLogUsecase
}

func NewAuthHandler(a *usecase.AuthUsecase, p *usecase.ProfileUsecase, al *usecase.AuditLogUsecase) *AuthHandler {
	return &AuthHandler{auth: a, profile: p, auditLog: al}
}

// @Summary		Cek status NIM
// @Description	Memeriksa apakah NIM sudah terdaftar (langkah 1 first-time login)
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.CekNIMRequest	true	"Payload NIM"
// @Success		200		{object}	response.Envelope{data=dto.CekNIMResponse}
// @Failure		400		{object}	response.Envelope
// @Router			/praktikum/auth/cek-nim [post]
func (h *AuthHandler) CekNIM(c *gin.Context) {
	var req dto.CekNIMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "NIM wajib diisi", err.Error())
		return
	}
	res, err := h.auth.CekNIM(req.NIM)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Status NIM", res)
}

// @Summary		Login Pengguna
// @Description	Login menggunakan NIM dan Password
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.LoginRequest	true	"Payload Login"
// @Success		200		{object}	response.Envelope{data=dto.AuthResponse}
// @Failure		400		{object}	response.Envelope
// @Router			/praktikum/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.auth.Login(req)
	if err != nil {
		// MEDIUM-01: pesan generik — user-ada vs user-tidak-ada vs password
		// salah semuanya balas 401 yang sama (mencegah enumerasi NIM/email).
		_ = h.auditLog.LogAction(0, req.Identifier, "LOGIN_FAILED", "Gagal login", middleware.ClientIP(c), c.Request.UserAgent())
		response.Fail(c, http.StatusUnauthorized, "NIM/email atau password salah", nil)
		return
	}
	_ = h.auditLog.LogAction(0, req.Identifier, "LOGIN", "Login berhasil", middleware.ClientIP(c), c.Request.UserAgent())
	response.OK(c, http.StatusOK, "Login berhasil", res)
}

// @Summary		Registrasi Mahasiswa
// @Description	Mendaftarkan password untuk pertama kali bagi mahasiswa yang sudah memiliki NIM terdaftar
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RegisterRequest	true	"Payload Register"
// @Success		201		{object}	response.Envelope{data=dto.AuthResponse}
// @Failure		400		{object}	response.Envelope
// @Router			/praktikum/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Password minimal 6 karakter", err.Error())
		return
	}
	res, err := h.auth.Register(req)
	if err != nil {
		_ = h.auditLog.LogAction(0, req.NIM, "REGISTER_FAILED", "Gagal registrasi: "+err.Error(), middleware.ClientIP(c), c.Request.UserAgent())
		mapError(c, err)
		return
	}
	_ = h.auditLog.LogAction(0, req.NIM, "REGISTER", "Registrasi akun berhasil", middleware.ClientIP(c), c.Request.UserAgent())
	response.Created(c, "Registrasi berhasil", res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.UserID(c)
	_ = h.auditLog.LogAction(userID, "", "LOGOUT", "Logout berhasil", middleware.ClientIP(c), c.Request.UserAgent())
	h.auth.Logout(userID)
	response.OK(c, http.StatusOK, "Logout berhasil", nil)
}

// Selalu 200 (anti-enumeration) tetapi membawa status pengiriman sebenarnya
// supaya pengguna yang akunnya belum punya email tidak menunggu kode kosong.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Email tidak valid", err.Error())
		return
	}
	hasil := h.auth.ForgotPassword(req.Email)
	pesan := "Jika email terdaftar, kode OTP telah dikirim. Cek inbox Anda."
	if !hasil.Sent && hasil.Reason == "no_email_channel" {
		pesan = "Akun ini belum punya email untuk menerima kode OTP. Isi email dulu di halaman Profil, lalu coba lagi."
	} else if !hasil.Sent {
		pesan = "Kode OTP gagal dikirim. Coba lagi beberapa saat lagi."
	}
	response.OK(c, http.StatusOK, pesan, hasil)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordViaOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	if err := h.auth.ResetPassword(req); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Password berhasil direset. Silakan login.", nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.profile.Get(middleware.UserID(c))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Profil", u)
}
