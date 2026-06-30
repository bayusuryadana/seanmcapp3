package bootstrap

import (
	"errors"
	"log"
	"net/http"
	"os"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter(mainServices MainServices, walletSettings util.WalletSettings) *gin.Engine {
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
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		wallet := api.Group("/wallet")
		{
			wallet.POST("/login", func(c *gin.Context) {
				var body struct {
					Password string `json:"password"`
				}
				if err := c.ShouldBindJSON(&body); err != nil {
					c.String(http.StatusBadRequest, "Invalid request")
					return
				}
				token := util.JwtCreateToken(walletSettings, body.Password)
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

			wallet.DELETE("/delete/:id", authMiddleware(walletSettings), func(c *gin.Context) {
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

			stock.DELETE("/delete/:id", authMiddleware(walletSettings), func(c *gin.Context) {
				name := c.Param("id")
				res, err := mainServices.StockService.Delete(name)
				resolve(c, res, err)
			})
		}

		instagram := api.Group("/instagram")
		{
			instagram.GET("/trigger", func(c *gin.Context) {
				go safeRun(mainServices.InstagramService.Run)
				c.JSON(http.StatusOK, gin.H{"data": "Instagram fetch triggered"})
			})
		}
	}

	return r
}

func safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] recovered from panic in background job: %v", r)
		}
	}()
	fn()
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
