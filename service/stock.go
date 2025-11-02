package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"

	"github.com/tidwall/gjson"
)

type StockService interface {
	Run()

	GetAll() ([]DashboardStock, error)
	Create(stock DashboardStock) (string, error)
	Update(stock DashboardStock) (string, error)
	Delete(name string) (string, error)
}

type StockServiceImpl struct {
	StockRepo      repository.StockRepo
	TelegramClient external.TelegramClient
	ChatId         int64
}

var urlTemplate = "https://query1.finance.yahoo.com/v8/finance/chart/{{name}}.jk"

func (s *StockServiceImpl) Run() {
	stocks, err := s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve data from DB: %v\n", err)
		return
	}

	for _, stock := range stocks {
		stockUrl := strings.NewReplacer(
			"{{name}}", stock.Name,
		).Replace(urlTemplate)

		resp, err := http.Get(stockUrl)
		if err != nil {
			log.Printf("[ERROR] cannot fetch stock data: %v\n", err)
			continue
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[ERROR] reading response: %v\n", err)
			continue
		}

		jsonStr := string(body)
		regularMarketPrice := gjson.Get(jsonStr, "chart.result.0.meta.regularMarketPrice")
		if regularMarketPrice.Exists() {
			currentPrice := regularMarketPrice.Int()
			updatedStock := repository.Stock{
				Name:         stock.Name,
				BestPrice:    stock.BestPrice,
				CurrentPrice: &currentPrice,
				FairPrice:    stock.FairPrice,
				Status:       stock.Status,
			}
			_, err := s.StockRepo.Update(updatedStock)
			if err != nil {
				log.Printf("[ERROR] cannot update stock: %v\n", err)
				continue
			}
		} else {
			log.Printf("[ERROR] stock %s not found in json\n", stock.Name)
		}
	}

	var result []string
	for _, stock := range stocks {
		if stock.CurrentPrice == nil || stock.FairPrice == nil || stock.BestPrice == nil {
			continue
		}

		// status = 0 and current_price <= best_price
		if stock.Status == false && stock.BestPrice != nil && *stock.CurrentPrice <= *stock.BestPrice {
			result = append(result, fmt.Sprintf("%s hitting best price", stock.Name))
		}

		// status = 1 and current_price >= fair_price
		if stock.Status == true && stock.FairPrice != nil && *stock.CurrentPrice >= *stock.FairPrice {
			result = append(result, fmt.Sprintf("%s reaching fair price", stock.Name))
		}
	}

	if len(result) > 0 {
		log.Println("[INFO] stocks hit/reach")
		finalResult := strings.Join(result, "\n")
		_, err := s.TelegramClient.SendMessage(s.ChatId, finalResult)
		if err != nil {
			log.Printf("[ERROR] cannot send message for the final result: %v\n", err)
		}
	}
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
	st := repository.Stock(stock)
	return s.StockRepo.Create(st)
}

func (s *StockServiceImpl) Update(stock DashboardStock) (string, error) {
	st := repository.Stock(stock)
	return s.StockRepo.Update(st)
}

func (s *StockServiceImpl) Delete(name string) (string, error) {
	return s.StockRepo.Delete(name)
}

type DashboardStock struct {
	Name         string `json:"name"`
	BestPrice    *int64 `json:"best_price,omitempty"`
	CurrentPrice *int64 `json:"current_price,omitempty"`
	FairPrice    *int64 `json:"fair_price,omitempty"`
	Status       bool   `json:"status"`
}
