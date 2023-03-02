package http

import (
	"github.com/22Fariz22/gophermart/pkg/logger"
	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Handler struct {
	useCase auth.UseCase
	l       logger.Interface
}

func NewHandler(useCase auth.UseCase, l logger.Interface) *Handler {
	return &Handler{
		useCase: useCase,
		l:       l,
	}
}

type signInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) SignUp(c *gin.Context) {
	inp := new(signInput)

	if err := c.BindJSON(inp); err != nil {
		h.l.Info("err BindJSON.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := h.useCase.SignUp(c.Request.Context(), h.l, inp.Login, inp.Password); err != nil {
		if err == auth.ErrLoginIsAlreadyTaken {
			h.l.Info("Err Login Is Already Taken")
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		if err == auth.ErrBadRequest {
			h.l.Info("Bad Request.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		h.l.Info("Internal Server Error.")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	token, err := h.useCase.SignIn(c.Request.Context(), h.l, inp.Login, inp.Password)
	if err != nil {
		if err == auth.ErrUserNotFound {
			h.l.Info("User Not Found.")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	c.Header("Authorization", token)

	c.Status(http.StatusOK)
}

func (h *Handler) SignIn(c *gin.Context) {
	log.Println("auth-handler-sugnIn().")
	inp := new(signInput)

	if err := c.BindJSON(inp); err != nil {
		h.l.Info("err BindJSON.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := h.useCase.SignIn(c.Request.Context(), h.l, inp.Login, inp.Password)
	if err != nil {
		if err == auth.ErrUserNotFound {
			h.l.Info("User Not Found.")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	c.Header("Authorization", token)
}
