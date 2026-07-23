package bootstrap

import (
	"log"
	"net/http"
	"seanmcapp/util"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
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

func InitScheduler(mainServices MainServices) *cron.Cron {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(loc),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	schedulers := []*Scheduler{
		{Task: mainServices.NewsService, CronExpr: "0 0 9 * * *", Repeat: true},
		{Task: mainServices.StockService, CronExpr: "0 0 19 * * *", Repeat: true},
		{Task: mainServices.InstagramService, CronExpr: "0 0 * * * *", Repeat: true},
	}

	for _, s := range schedulers {
		_, err := s.Schedule(c)
		if err != nil {
			log.Fatalf("Failed to schedule task: %v", err)
		}
	}

	log.Println("Running scheduled jobs...")
	c.Start()
	return c
}
