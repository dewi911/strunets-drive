package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strunetsdrive/internal/models"
)

type AuthHandler struct {
	authService UserService
}

func NewAuthHandler(authService UserService) *AuthHandler {
	return &AuthHandler{authService}
}

func (a *AuthHandler) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	auth := r.Group("/auth").Use(middlewares...)
	{
		auth.POST("/sign-up", a.signUp)
		auth.POST("/login", a.login)
		auth.GET("/refresh", a.refresh)
	}

}

func (a *AuthHandler) signUp(ctx *gin.Context) {
	var input models.SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authService.SingUp(ctx, input)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User successfully registered"})
}

func (a *AuthHandler) login(ctx *gin.Context) {
	var inp models.LoginInput
	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("validation request body error", err))
		return
	}

	accessToken, refreshToken, err := a.authService.Login(ctx, inp)
	if err != nil {
		if errors.Is(err, errors.New("user with such credentials not found")) { //TODO REPLACE ERROR.new TO CONST
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid credentials"})
		return
	}

	ctx.SetCookie("refresh-token", refreshToken, 3600, "/auth", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func (a *AuthHandler) refresh(ctx *gin.Context) {
	cookie, err := ctx.Cookie("refresh-token")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("get cookie from request error", err))
		return
	}

	accessToken, refreshToken, err := a.authService.RefreshSession(ctx, cookie)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("refresh token error", err))
		return
	}

	ctx.SetCookie("refresh-token", refreshToken, 3600, "/auth", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}
