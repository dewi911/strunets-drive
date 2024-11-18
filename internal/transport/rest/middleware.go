package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const (
	ctxUserIDKey     = "user-id"
	ctxUserRoleIDKey = "user-role-id"
	ctxUsernameKey   = "username" // Добавляем новую константу
)
const AuthorizationHeaderName = "Authorization"

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fields := logrus.Fields{
			"method":          c.Request.Method,
			"uri":             c.Request.RequestURI,
			"request-in-time": t.Format(time.RFC3339),
		}

		c.Next()

		dur := time.Since(t)
		fields["request-handling-duration"] = dur.Milliseconds()

		logrus.WithFields(fields).Info()
	}
}

func (a *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getTokenFromRequest(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("token is missing or invalid", err))
			return
		}

		// Обновляем сигнатуру метода ParseToken для получения username
		userID, username, err := a.authService.ParseToken(c, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("wrong authorization token", err))
			return
		}

		c.Set(ctxUserIDKey, userID)
		//c.Set(ctxUserRoleIDKey, roleID)
		c.Set(ctxUsernameKey, username) // Устанавливаем username в контекст

		c.Next()
	}
}

// Вспомогательная функция для получения username из контекста
func GetUsernameFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get(ctxUsernameKey)
	if !exists {
		return "", errors.New("username not found in context")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", errors.New("username is of invalid type")
	}

	return usernameStr, nil
}

func getTokenFromRequest(c *gin.Context) (string, error) {
	header := c.GetHeader(AuthorizationHeaderName)
	if header == "" {
		return "", errors.Wrap(nil, "empty authorization header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.Wrap(nil, "invalid authorization header")
	}

	if len(headerParts[1]) == 0 {
		return "", errors.Wrap(nil, "token is empty")
	}

	return headerParts[1], nil
}
