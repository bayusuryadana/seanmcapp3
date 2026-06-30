package bootstrap

import (
	"io"
	"log"
	"net/http"
	"os"
	"seanmcapp/util"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter(mainServices MainServices, walletSettings util.WalletSettings) {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080", "https://seanmcapp.herokuapp.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
				token := util.JwtCreateToken(walletSettings, userPassword)
				if token == "" {
					c.String(http.StatusUnauthorized, "Invalid password")
				} else {
					c.String(http.StatusOK, token)
				}
			})

			wallet.GET("/dashboard", authMiddleware(walletSettings), func(c *gin.Context) {
				dateStr := c.Query("date")
				date, _ := strconv.Atoi(dateStr)
				res, err := mainServices.WalletService.Dashboard(date)
				resolve(c, res, err)
			})

			wallet.POST("/create", authMiddleware(walletSettings), handleJSON(mainServices.WalletService.Create))
			wallet.POST("/update", authMiddleware(walletSettings), handleJSON(mainServices.WalletService.Update))

			wallet.GET("/delete/:id", authMiddleware(walletSettings), func(c *gin.Context) {
				idStr := c.Param("id")
				id, _ := strconv.Atoi(idStr)
				res, err := mainServices.WalletService.Delete(id)
				resolve(c, res, err)
			})
		}

		stock := api.Group("/stock")
		{
			stock.POST("/getAll", authMiddleware(walletSettings), func(c *gin.Context) {
				res, err := mainServices.StockService.GetAll()
				resolve(c, res, err)
			})

			stock.POST("/refresh", authMiddleware(walletSettings), func(c *gin.Context) {
				res, err := mainServices.StockService.RefreshPrices()
				resolve(c, res, err)
			})

			stock.POST("/create", authMiddleware(walletSettings), handleJSON(mainServices.StockService.Create))
			stock.POST("/update", authMiddleware(walletSettings), handleJSON(mainServices.StockService.Update))

			stock.GET("/delete/:id", authMiddleware(walletSettings), func(c *gin.Context) {
				name := c.Param("id")
				res, err := mainServices.StockService.Delete(name)
				resolve(c, res, err)
			})
		}

		instagram := api.Group("/instagram")
		{
			instagram.GET("/trigger", func(c *gin.Context) {
				go mainServices.InstagramService.Run()
				c.JSON(http.StatusOK, gin.H{"data": "Instagram fetch triggered"})
			})
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
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

func telegramWebhookServiceReceive(payload string) any {
	return gin.H{"status": "received", "payload": payload}
}
