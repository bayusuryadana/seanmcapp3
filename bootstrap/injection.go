package bootstrap

import (
	"database/sql"
	"log"
	"net"
	"net/url"
	"seanmcapp/external"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"
	"time"

	_ "github.com/lib/pq"
)

type MainServices struct {
	BirthdayService  service.BirthdayService
	WalletService    service.WalletService
	NewsService      service.NewsService
	StockService     service.StockService
	InstagramService service.InstagramService
}

func GetMainServices(settings util.AppsSettings) (MainServices, *sql.DB) {

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(settings.DBSettings.User, settings.DBSettings.Pass),
		Host:     net.JoinHostPort(settings.DBSettings.Host, "5432"),
		Path:     settings.DBSettings.Name,
		RawQuery: "sslmode=require",
	}

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot reach database: %v", err)
	}

	// Keep the pool within Heroku Postgres connection limits.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	peopleRepo := &repository.PeopleRepoImpl{DB: db}
	walletRepo := &repository.WalletRepoImpl{DB: db}
	stockRepo := &repository.StockRepoImpl{DB: db}
	instagramAccountRepo := &repository.InstagramAccountRepoImpl{DB: db}

	telegramClient := external.NewTelegramClient(settings.TelegramSettings.Endpoint, settings.TelegramSettings.Botname)
	instagramClient := external.NewInstagramClient(settings.IGSettings.SessionID, settings.IGSettings.CSRFToken)
	stockClient := external.NewStockClient()

	birthdayService := &service.BirthdayServiceImpl{PeopleRepo: peopleRepo, TelegramClient: telegramClient, PersonalChatID: settings.TelegramSettings.PersonalChatID}
	walletService := &service.WalletServiceImpl{WalletRepo: walletRepo}
	newsService := service.NewNewsService(telegramClient, settings.TelegramSettings.GroupChatID)
	stockService := &service.StockServiceImpl{StockRepo: stockRepo, StockClient: stockClient, TelegramClient: telegramClient, PersonalChatID: settings.TelegramSettings.PersonalChatID}
	instagramService := &service.InstagramServiceImpl{InstagramAccountRepo: instagramAccountRepo, InstagramClient: instagramClient, TelegramClient: telegramClient, PersonalChatID: settings.TelegramSettings.PersonalChatID}

	return MainServices{
		BirthdayService:  birthdayService,
		WalletService:    walletService,
		NewsService:      newsService,
		StockService:     stockService,
		InstagramService: instagramService,
	}, db

}
