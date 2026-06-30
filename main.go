package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"seanmcapp/bootstrap"
	"seanmcapp/util"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, relying on system environment variables")
	}

	settings := util.GetAppSettings()

	mainServices, db := bootstrap.GetMainServices(settings)
	defer db.Close()

	cronScheduler := bootstrap.InitScheduler(mainServices)

	router := bootstrap.InitRouter(mainServices, settings.WalletSettings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}

	srv := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Println("server started on :" + port)

	// Wait for SIGTERM (Heroku dyno restart) or SIGINT (Ctrl-C).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	// Let in-flight cron jobs finish (bounded by the same deadline) so a
	// background job isn't severed mid-write on a Heroku restart.
	select {
	case <-cronScheduler.Stop().Done():
	case <-shutdownCtx.Done():
		log.Println("cron jobs did not finish before shutdown deadline")
	}
}
