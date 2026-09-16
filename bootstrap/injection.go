package bootstrap

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"net/url"
	"seanmcapp/external"
	"seanmcapp/repository"
	"seanmcapp/service"
	"seanmcapp/util"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	_ "github.com/lib/pq"
)

type MainServices struct {
	WalletService         service.WalletService
	NewsService           service.NewsService
	StockService          service.StockService
	InstagramService      service.InstagramService
	TelegramUpdateHandler service.TelegramUpdateHandler
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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	walletRepo := &repository.WalletRepoImpl{DB: db}
	stockRepo := &repository.StockRepoImpl{DB: db}
	instagramAccountRepo := &repository.InstagramAccountRepoImpl{DB: db}

	httpClient := &http.Client{Timeout: external.HTTPTimeout}
	telegramClient := &external.TelegramClientImpl{
		Endpoint:     settings.TelegramSettings.Endpoint,
		Botname:      settings.TelegramSettings.Botname,
		Client:       httpClient,
		UploadClient: &http.Client{Timeout: external.TelegramUploadTimeout},
	}
	instagramHTTPClient, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithTimeoutSeconds(15),
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithNotFollowRedirects(),
	)
	if err != nil {
		log.Fatalf("cannot initialize Instagram client: %v", err)
	}
	instagramClient := &external.InstagramClientImpl{Client: instagramHTTPClient}
	stockClient := &external.StockClientImpl{Client: httpClient}

	walletService := &service.WalletServiceImpl{WalletRepo: walletRepo}
	newsService := &service.NewsServiceImpl{
		TelegramClient: telegramClient,
		GroupChatID:    settings.TelegramSettings.GroupChatID,
		HTTPClient:     httpClient,
		Sources: []service.NewsObject{
			service.Detik{},
			service.Tirtol{},
			service.Kumparan{},
			service.CNA{},
			service.Mothership{},
			service.Reuters{},
		},
	}
	stockService := &service.StockServiceImpl{StockRepo: stockRepo, StockClient: stockClient, TelegramClient: telegramClient, PersonalChatID: settings.TelegramSettings.PersonalChatID}
	instagramService := &service.InstagramServiceImpl{InstagramAccountRepo: instagramAccountRepo, InstagramClient: instagramClient, TelegramClient: telegramClient, PersonalChatID: settings.TelegramSettings.PersonalChatID}
	telegramUpdateHandler := &service.TelegramCommandDispatcher{
		Botname:  settings.TelegramSettings.Botname,
		Commands: nil,
	}

	return MainServices{
		WalletService:         walletService,
		NewsService:           newsService,
		StockService:          stockService,
		InstagramService:      instagramService,
		TelegramUpdateHandler: telegramUpdateHandler,
	}, db

}
