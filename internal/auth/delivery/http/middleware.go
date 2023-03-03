package http

import (
	"net/http"
	"strings"

	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	usecase auth.UseCase
	l       logger.Interface
}

func NewAuthMiddleware(usecase auth.UseCase, l logger.Interface) gin.HandlerFunc {
	return (&AuthMiddleware{
		usecase: usecase,
		l:       l,
	}).Handle
}

func (m *AuthMiddleware) Handle(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		m.l.Info("authHeader == ''.Status Unauthorized.")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	headerParts := strings.Split(authHeader, " ")

	var token string

	if headerParts[0] == "Bearer" && len(headerParts) == 2 {
		token = headerParts[1]
	} else if len(headerParts) == 1 {
		token = headerParts[0]
	}

	user, err := m.usecase.ParseToken(c.Request.Context(), m.l, token)
	if err != nil {
		status := http.StatusInternalServerError
		if err == auth.ErrInvalidAccessToken {
			m.l.Info("err == auth.ErrInvalidAccessToken.Invalid Access Token.")
			status = http.StatusUnauthorized
		}
		m.l.Info("Status Internal Server Error")
		c.AbortWithStatus(status)
		return
	}
	c.Set(auth.CtxUserKey, user)
}
