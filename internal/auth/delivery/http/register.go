package http

import (
	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, uc auth.UseCase, l logger.Interface) {
	h := NewHandler(uc, l)

	authEndpoints := router.Group("/api/user")
	{
		authEndpoints.POST("/register", h.SignUp)
		authEndpoints.POST("/login", h.SignIn)
	}
}
