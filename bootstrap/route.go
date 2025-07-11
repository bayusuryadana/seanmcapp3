package bootstrap

import (
	"io"
	"net/http"
	"os"
	"seanmcapp/service"
	"seanmcapp/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

func InitRouter(mainServices MainServices) {
	r := gin.Default()

	// Frontend routes
	r.GET("/", serveIndex)
	r.Static("/static", util.GetFrontendPath()+"/static")
	r.NoRoute(serveIndex)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/webhook", func(c *gin.Context) {
			body, _ := io.ReadAll(c.Request.Body)
			result := telegramWebhookServiceReceive(string(body))
			c.JSON(http.StatusOK, result)
		})

		wallet := api.Group("/wallet")
		{
			wallet.GET("/login/", func(c *gin.Context) {
				c.String(http.StatusUnauthorized, "Invalid password")
			})

			wallet.GET("/login/:password", func(c *gin.Context) {
				userPassword := c.Param("password")
				token := util.JwtCreateToken(util.GetAppSettings().WalletSettings, userPassword)
				if token == "" {
					c.String(http.StatusUnauthorized, "Invalid password")
				} else {
					c.String(http.StatusOK, token)
				}
			})

			wallet.GET("/dashboard", authMiddleware(), func(c *gin.Context) {
				dateStr := c.Query("date")
				date, _ := strconv.Atoi(dateStr)
				res, err := mainServices.WalletService.Dashboard(date)
				resolve(c, res, err)
			})

			wallet.POST("/create", authMiddleware(), func(c *gin.Context) {
				var payload service.DashboardWallet
				if err := c.ShouldBindJSON(&payload); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}

				res, err := mainServices.WalletService.Create(payload)
				resolve(c, res, err)
			})

			wallet.POST("/update", authMiddleware(), func(c *gin.Context) {
				var payload service.DashboardWallet
				if err := c.ShouldBindJSON(&payload); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}
				res, err := mainServices.WalletService.Update(payload)
				resolve(c, res, err)
			})

			wallet.GET("/delete/:id", authMiddleware(), func(c *gin.Context) {
				idStr := c.Param("id")
				id, _ := strconv.Atoi(idStr)
				res, err := mainServices.WalletService.Delete(id)
				resolve(c, res, err)
			})
		}
	}

	r.Run(":8080")
}

// Auth Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if !util.JwtValidateToken(util.GetAppSettings().WalletSettings, token) {
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Temp function
func telegramWebhookServiceReceive(payload string) any {
	return gin.H{"status": "received", "payload": payload}
}
