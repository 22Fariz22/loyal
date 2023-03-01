package http

import (
	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/history"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Handler struct {
	useCase history.UseCase
	l       logger.Interface
}

func NewHandler(useCase history.UseCase, l logger.Interface) *Handler {
	return &Handler{
		useCase: useCase,
		l:       l,
	}
}

type BalanceResponce struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h *Handler) GetBalance(c *gin.Context) {
	user := c.MustGet(auth.CtxUserKey).(*entity.User)

	u, err := h.useCase.GetBalance(c.Request.Context(), h.l, user)
	if err != nil {
		h.l.Error("Status Internal ServerError: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	br := toBalanceResponce(u)

	c.JSON(http.StatusOK, BalanceResponce{
		Current:   br.Current,
		Withdrawn: br.Withdrawn,
	})
}

func toBalanceResponce(u *entity.User) *BalanceResponce {
	return &BalanceResponce{
		Current:   float64(u.BalanceTotal) / 100,
		Withdrawn: float64(u.WithdrawTotal) / 100,
	}
}

type InputWithdraw struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *Handler) Withdraw(c *gin.Context) {
	user := c.MustGet(auth.CtxUserKey).(*entity.User)

	inp := new(InputWithdraw)
	if err := c.BindJSON(inp); err != nil {
		h.l.Error("history-handler-Withdraw()-BindJSON-err: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err := h.useCase.Withdraw(c.Request.Context(), h.l, user, inp.Order, int(inp.Sum*100))
	if err != nil {
		if err == history.ErrNotEnoughFunds { //если не достаточно баллов
			h.l.Info("Not Enough Funds")
			c.AbortWithStatus(http.StatusPaymentRequired)
		}

		h.l.Error("Status Internal Server Error: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type HistoryResp struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *Handler) InfoWithdrawal(c *gin.Context) {
	user := c.MustGet(auth.CtxUserKey).(*entity.User)

	hist, err := h.useCase.InfoWithdrawal(c.Request.Context(), h.l, user)
	if err != nil {
		if err == history.ErrThereIsNoWithdrawal { //нет списаний
			h.l.Error("There Is No Withdrawal")
			c.AbortWithStatus(http.StatusNoContent)
		}
		h.l.Error("")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, toHistoryResps(hist))
}

func toHistoryResps(eh []*entity.History) []*HistoryResp {
	out := make([]*HistoryResp, len(eh))

	for i, o := range eh {
		out[i] = toHistoryResp(o)
	}

	return out
}

func toHistoryResp(h *entity.History) *HistoryResp {
	strTime := h.ProcessedAt.Format(time.RFC3339)

	return &HistoryResp{
		Order:       h.Number,
		Sum:         float64(h.Sum) / 100,
		ProcessedAt: strTime,
	}
}
