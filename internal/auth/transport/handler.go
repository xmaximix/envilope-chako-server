package transport

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/xmaximix/envilope-chako-server/internal/auth/repository"
	"github.com/xmaximix/envilope-chako-server/internal/auth/usecase"
	"github.com/xmaximix/envilope-chako-server/internal/config"
	"github.com/xmaximix/envilope-chako-server/pkg/email"
	"go.uber.org/zap"
	"net/http"
)

func RegisterAuthRoutes(
	r *gin.Engine,
	db *sqlx.DB,
	ac config.AuthConfig,
	logg *zap.SugaredLogger,
) {
	userRepo := repository.NewUserRepo(db)
	emailSender := email.NewSMTPSender(ac.SMTPHost, ac.SMTPPort, ac.SMTPUser, ac.SMTPPass, ac.SMTPFrom)
	regUC := usecase.NewRegisterUser(userRepo, emailSender)
	loginUC := usecase.NewLoginUser(userRepo, []byte(ac.JWTSecret), ac.AccessTokenTTL, ac.RefreshTokenTTL)
	refreshUC := usecase.NewRefreshToken(userRepo, []byte(ac.JWTSecret), ac.AccessTokenTTL)

	rate, _ := limiter.NewRateFromFormatted("5-M")
	store := memory.NewStore()
	loginLimiter := limitergin.NewMiddleware(limiter.New(store, rate))

	grp := r.Group("/auth")
	{
		grp.POST("/register", func(c *gin.Context) {
			var req struct {
				Email    string `json:"email"    binding:"required,email"`
				Password string `json:"password" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": "BAD_REQUEST",
					"msg":  err.Error(),
				})
				return
			}
			if err := regUC.Execute(c, req.Email, req.Password); err != nil {
				c.Error(AppError{"EMAIL_EXISTS", http.StatusConflict, err.Error()})
				return
			}
			c.Status(http.StatusAccepted)
		})

		grp.POST("/login", loginLimiter, func(c *gin.Context) {
			var req struct {
				Email, Password string `json:"email" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.Error(AppError{"BAD_REQUEST", http.StatusBadRequest, err.Error()})
				return
			}
			tokens, err := loginUC.Execute(c, req.Email, req.Password)
			if err != nil {
				c.Error(AppError{"AUTH_FAILED", http.StatusUnauthorized, err.Error()})
				return
			}
			c.JSON(http.StatusOK, tokens)
		})

		grp.POST("/refresh", func(c *gin.Context) {
			var req struct {
				Token string `json:"refresh_token" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.Error(AppError{"BAD_REQUEST", http.StatusBadRequest, err.Error()})
				return
			}
			at, err := refreshUC.Execute(c, req.Token)
			if err != nil {
				c.Error(AppError{"INVALID_REFRESH", http.StatusUnauthorized, err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"access_token": at})
		})
	}
}
