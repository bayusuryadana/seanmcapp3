package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"seanmcapp/external"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"

	_ "github.com/lib/pq"
)

type MainServices struct {
	WarmupDBService   service.WarmupDBService
	BirthdayService   service.BirthdayService
	WalletService     service.WalletService
	NewsService       service.NewsService
	StockService      service.StockService
	InstagramService  service.InstagramService
}

func GetMainServices(settings util.AppsSettings) (MainServices, *sql.DB) {

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		settings.DBSettings.Host, 5432, settings.DBSettings.User, settings.DBSettings.Pass, settings.DBSettings.Name, "require",
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	peopleRepo := &repository.PeopleRepoImpl{DB: db}
	walletRepo := &repository.WalletRepoImpl{DB: db}
	stockRepo := &repository.StockRepoImpl{DB: db}
	instagramAccountRepo := &repository.InstagramAccountRepoImpl{DB: db}

	telegramClient := &external.TelegramClientImpl{Endpoint: settings.TelegramSettings.Endpoint, Botname: settings.TelegramSettings.Botname}
	instagramClient := external.NewInstagramClient(settings.IGSettings.SessionID, settings.IGSettings.CSRFToken)

	warmupDBService := &service.WarmupDBServiceImpl{PeopleRepo: peopleRepo}
	birthdayService := &service.BirthdayServiceImpl{PeopleRepo: peopleRepo, TelegramClient: telegramClient}
	walletService := &service.WalletServiceImpl{WalletRepo: walletRepo, StockRepo: stockRepo}
	newsService := &service.NewsServiceImpl{TelegramClient: telegramClient}
	stockService := &service.StockServiceImpl{StockRepo: stockRepo, TelegramClient: telegramClient}
	instagramService := &service.InstagramServiceImpl{InstagramAccountRepo: instagramAccountRepo, InstagramClient: instagramClient, TelegramClient: telegramClient}

	return MainServices{
		WarmupDBService:  warmupDBService,
		BirthdayService:  birthdayService,
		WalletService:    walletService,
		NewsService:      newsService,
		StockService:     stockService,
		InstagramService: instagramService,
	}, db

}
