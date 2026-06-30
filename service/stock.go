package service

import (
	"errors"
	"fmt"
	"log"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"
)

type StockService interface {
	Run()
	RefreshPrices() ([]DashboardStock, error)

	GetAll() ([]DashboardStock, error)
	Create(stock DashboardStock) (string, error)
	Update(stock DashboardStock) (string, error)
	Delete(name string) (string, error)
}

type StockServiceImpl struct {
	StockRepo      repository.StockRepo
	StockClient    external.StockClient
	TelegramClient external.TelegramClient
	PersonalChatID int64
}

func (s *StockServiceImpl) Run() {
	stocks, err := s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve data from DB: %v\n", err)
		return
	}

	s.fetchAndUpdatePrices(stocks)
	stocks, err = s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve refreshed data from DB: %v\n", err)
		return
	}

	var result []string
	for _, stock := range stocks {
		if stock.CurrentPrice == nil {
			continue
		}

		// status = 0 and current_price <= best_price
		if stock.Status == false && *stock.CurrentPrice <= stock.BestPrice {
			result = append(result, fmt.Sprintf("%s hitting best price", stock.Name))
		}

		// status = 1 and current_price >= fair_price
		if stock.Status == true && *stock.CurrentPrice >= stock.FairPrice {
			result = append(result, fmt.Sprintf("%s reaching fair price", stock.Name))
		}
	}

	if len(result) > 0 {
		log.Println("[INFO] stocks hit/reach")
		finalResult := strings.Join(result, "\n")
		_, err := s.TelegramClient.SendMessage(s.PersonalChatID, finalResult)
		if err != nil {
			log.Printf("[ERROR] cannot send message for the final result: %v\n", err)
		}
	}
}

func (s *StockServiceImpl) fetchAndUpdatePrices(stocks []DashboardStock) {
	for _, stock := range stocks {
		currentPrice, err := s.StockClient.GetPrice(stock.Name)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			continue
		}

		updatedStock := repository.Stock{
			Name:         stock.Name,
			BestPrice:    stock.BestPrice,
			CurrentPrice: &currentPrice,
			FairPrice:    stock.FairPrice,
			Status:       stock.Status,
			BuyPrice:     stock.BuyPrice,
			Lot:          stock.Lot,
		}
		if _, err := s.StockRepo.Update(updatedStock); err != nil {
			log.Printf("[ERROR] cannot update stock: %v\n", err)
			continue
		}
	}
}

func (s *StockServiceImpl) RefreshPrices() ([]DashboardStock, error) {
	stocks, err := s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve data from DB: %v\n", err)
		return nil, err
	}

	s.fetchAndUpdatePrices(stocks)

	return s.GetAll()
}

func (s *StockServiceImpl) GetAll() ([]DashboardStock, error) {
	stocks, err := s.StockRepo.GetAll()
	var dashboardStocks []DashboardStock
	for _, st := range stocks {
		dashboardStocks = append(dashboardStocks, DashboardStock(st))
	}
	return dashboardStocks, err
}

func (s *StockServiceImpl) Create(stock DashboardStock) (string, error) {
	if stock.BestPrice <= 0 || stock.FairPrice <= 0 {
		return "", errors.New("best_price and fair_price are required and must be > 0")
	}
	st := repository.Stock(stock)
	return s.StockRepo.Create(st)
}

func (s *StockServiceImpl) Update(stock DashboardStock) (string, error) {
	if stock.BestPrice <= 0 || stock.FairPrice <= 0 {
		return "", errors.New("best_price and fair_price are required and must be > 0")
	}
	st := repository.Stock(stock)
	return s.StockRepo.Update(st)
}

func (s *StockServiceImpl) Delete(name string) (string, error) {
	return s.StockRepo.Delete(name)
}

type DashboardStock struct {
	Name         string `json:"name"`
	BestPrice    int64  `json:"best_price"`
	CurrentPrice *int64 `json:"current_price,omitempty"`
	FairPrice    int64  `json:"fair_price"`
	Status       bool   `json:"status"`
	BuyPrice     *int64 `json:"buy_price,omitempty"`
	Lot          *int64 `json:"lot,omitempty"`
}

