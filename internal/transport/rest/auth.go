package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strunetsdrive/internal/models"
)

type Auth struct {
	authService UserService
}

func NewAuth(authService UserService) *Auth {
	return &Auth{authService}
}

func (a *Auth) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	auth := r.Group("/auth").Use(middlewares...)
	{
		auth.POST("/sign-up", a.signUp)
		auth.POST("/sign-in", a.login)
		auth.GET("/refresh", a.refresh)
	}

}

func (a *Auth) signUp(ctx *gin.Context) {
	var input models.SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("validation request body error", err))
		return
	}

	err := a.authService.SingUp(ctx, input)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("signing up error", err))
		return
	}

	ctx.JSON(http.StatusOK, input)
}

func (a *Auth) login(ctx *gin.Context) {
	var inp models.LoginInput
	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("validation request body error", err))
		return
	}

	accessToken, refreshToken, err := a.authService.Login(ctx, inp)
	if err != nil {
		if errors.Is(err, errors.New("user with such credentials not found")) { //TODO REPLACE ERROR.new TO CONST
			ctx.AbortWithStatusJSON(http.StatusNotFound, NewNotFoundError("user not found", err))
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("signing in error", err))
		return
	}

	ctx.SetCookie("refresh-token", refreshToken, 3600, "/auth", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func (a *Auth) refresh(ctx *gin.Context) {
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
