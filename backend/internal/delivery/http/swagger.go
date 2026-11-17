package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/docs"
)

//	@title			Lab-AP API
//	@version		1.0
//	@description	REST API Documentation for Web Lab-AP v3
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Developer
//	@contact.email	naufalraihans@example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api

// MEDIUM-02: Swagger membocorkan endpoint termasuk seluruh /admin/*
// (inject jawaban, konfigurasi, bulk delete). Di produksi TIDAK dipasang.
func registerSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
