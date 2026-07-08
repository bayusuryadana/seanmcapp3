package bootstrap

import (
	"errors"
	"net/http"
	"os"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"

	"github.com/gin-gonic/gin"
)


// Auth Middleware
func authMiddleware(walletSettings util.WalletSettings) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			// Let preflight through
			c.AbortWithStatus(http.StatusOK)
			return
		}

		token := c.GetHeader("Authorization")
		if !util.JwtValidateToken(walletSettings, token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}

// Frontend handler
func serveIndex(c *gin.Context) {
	content, err := os.ReadFile(util.GetFrontendPath() + "/index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func resolve[T any](c *gin.Context, result T, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"data": result})
		return
	}

	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func handleJSON[Req any, Res any](fn func(Req) (Res, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload Req
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		res, err := fn(payload)
		resolve(c, res, err)
	}
}
