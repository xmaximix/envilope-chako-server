package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppError struct {
	Code    string
	HTTP    int
	Message string
}

func (e AppError) Error() string { return e.Message }

func ErrorMiddleware(logg *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		if ae, ok := err.(AppError); ok {
			c.JSON(ae.HTTP, gin.H{"code": ae.Code, "msg": ae.Message})
		} else {
			logg.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "msg": "internal error"})
		}
	}
}
