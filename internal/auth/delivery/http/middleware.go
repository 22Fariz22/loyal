package http

import (
	"github.com/22Fariz22/loyal/internal/auth"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
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
	//cookieHeader := c.GetHeader("Cookie")

	log.Println("middleware c.GetHeader('Authorization'):", authHeader)
	//log.Println("middleware c.GetHeader('Cookie'):", cookieHeader)

	if authHeader == "" {
		m.l.Info("authHeader == ''.Status Unauthorized.")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	headerParts := strings.Split(authHeader, " ")

	log.Println("middleware len(c.GetHeader('Authorization')) after split ' ':", len(headerParts))
	log.Println("middleware len(c.GetHeader('Authorization'))  headerParts[0] ' ':", headerParts[0])
	log.Println("middleware len(c.GetHeader('Authorization'))  len(headerParts[0]) ' ':", len(headerParts[0]))

	//if len(headerParts) != 2 {
	//	m.l.Info("len(headerParts) != 2.Status Unauthorized.")
	//	c.AbortWithStatus(http.StatusUnauthorized)
	//	return
	//}

	//if headerParts[0] != "Bearer" {
	//	m.l.Info("headerParts[0] != 'Bearer'.Not bearer.Status Unauthorized.")
	//	c.AbortWithStatus(http.StatusUnauthorized)
	//	return
	//}
	var token string

	if headerParts[0] == "Bearer" && len(headerParts) == 2 {
		token = headerParts[1]
	} else if len(headerParts) == 1 {
		token = headerParts[0]
	}

	//splitToken := strings.Split(authHeader, "Bearer ")
	//authHeader = strings.TrimSpace(splitToken[1])
	//log.Println("middlw-authHeader after split:", authHeader)

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
