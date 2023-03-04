package http

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/order"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/theplant/luhn"
)

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Handler struct {
	useCase order.UseCase
	l       logger.Interface
}

func NewHandler(useCase order.UseCase, l logger.Interface) *Handler {
	return &Handler{
		useCase: useCase,
		l:       l,
	}
}

type Number struct {
	number string
}

func (h *Handler) PushOrder(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)

	if err != nil {
		h.l.Error("Status Bad Request: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// проверка по формату номера заказа и по алгоритму Луна
	conv, err := strconv.Atoi(string(payload))
	if err != nil {
		h.l.Error("error in conv.Atoi.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	resLuhn := luhn.Valid(conv)
	if !resLuhn {
		h.l.Error("error in algorithm Luhn")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	user := c.MustGet(auth.CtxUserKey).(*entity.User)
	if err := h.useCase.PushOrder(c.Request.Context(), h.l, user, string(payload)); err != nil {
		if err == order.ErrNumberHasAlreadyBeenUploaded {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		if err == order.ErrNumberHasAlreadyBeenUploadedByAnotherUser {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		h.l.Error("Status Internal ServerError: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusAccepted)
}

func (h *Handler) GetOrders(c *gin.Context) {
	user := c.MustGet(auth.CtxUserKey).(*entity.User)

	orders, err := h.useCase.GetOrders(c.Request.Context(), h.l, user)
	if err != nil {
		if err == order.ErrThereIsNoOrders {
			c.AbortWithStatus(http.StatusNoContent)
			h.l.Error("there is no orders")
			return
		}
		h.l.Error("Status Internal ServerError: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, toOrders(orders))
}

func toOrders(os []*entity.Order) []*Order {
	out := make([]*Order, len(os))

	for i, o := range os {
		out[i] = toOrder(o)
	}

	return out
}

func toOrder(o *entity.Order) *Order {
	strTime := o.UploadedAt.Format(time.RFC3339)
	return &Order{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    float64(o.Accrual) / 100,
		UploadedAt: strTime,
	}
}
