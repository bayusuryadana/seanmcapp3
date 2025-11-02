package service

import (
	"seanmcapp/repository"
)

type StockService interface {
	GetAll() ([]DashboardStock, error)
	Create(stock DashboardStock) (string, error)
	Update(stock DashboardStock) (string, error)
	Delete(name string) (string, error)
}

type StockServiceImpl struct {
	StockRepo repository.StockRepo
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
	Name         string  `json:"name"`
	BestPrice    *string `json:"best_price,omitempty"`
	CurrentPrice *int64  `json:"current_price,omitempty"`
	FairPrice    *int64  `json:"fair_price,omitempty"`
	Status       bool    `json:"status"`
}
