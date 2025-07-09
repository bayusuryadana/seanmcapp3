package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"seanmcapp/external"
	"seanmcapp/repository"
	"seanmcapp/service"

	_ "github.com/lib/pq"
)

type MainServices struct {
	WarmupDBService service.WarmupDBService
	BirthdayService service.BirthdayService
	WalletService   service.WalletService
}

func GetMainServices(settings AppsSettings) (MainServices, *sql.DB) {

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

	telegramClient := &external.TelegramClientImpl{Endpoint: settings.TelegramSettings.Endpoint, Botname: settings.TelegramSettings.Botname}

	warmupDBService := &service.WarmupDBServiceImpl{PeopleRepo: peopleRepo}
	birthdayService := &service.BirthdayServiceImpl{PeopleRepo: peopleRepo, TelegramClient: telegramClient, ChatId: settings.TelegramSettings.PersonalChatID}
	walletService := &service.WalletServiceImpl{WalletRepo: walletRepo}

	return MainServices{
		WarmupDBService: warmupDBService,
		BirthdayService: birthdayService,
		WalletService:   walletService,
	}, db

}
