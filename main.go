package main

import (
	"log"
	"seanmcapp/bootstrap"
	"seanmcapp/util"

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

	bootstrap.InitScheduler(mainServices)
	bootstrap.InitRouter(mainServices, settings.WalletSettings)

}
