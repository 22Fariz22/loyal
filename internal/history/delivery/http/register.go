package http

import (
	"github.com/22Fariz22/loyal/internal/history"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.RouterGroup, uc history.UseCase, l logger.Interface) {
	h := NewHandler(uc, l)

	historyEndpoints := router.Group("/api/user")
	{
		historyEndpoints.GET("/balance", h.GetBalance)
		historyEndpoints.POST("/balance/withdraw", h.Withdraw)
		historyEndpoints.GET("/withdrawals", h.InfoWithdrawal)
	}
}
