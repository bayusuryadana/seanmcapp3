package service

import (
	"errors"
	"fmt"
	"log"
	"math"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"
	"time"
)

type StockService interface {
	Run()
	InitializeSnapshots()
	RefreshPrices() ([]DashboardStock, error)

	GetAll() ([]DashboardStock, error)
	GetJKSE() (JKSEQuote, error)
	GetSummary(period string) (DashboardSummary, error)
	GetProgression(period string) ([]DashboardProgressPoint, error)
	Create(stock DashboardStock) (string, error)
	Update(stock DashboardStock) (string, error)
	Delete(name string) (string, error)
}

type JKSEQuote struct {
	CurrentPrice  float64 `json:"current_price"`
	PreviousClose float64 `json:"previous_close"`
}

type StockServiceImpl struct {
	StockRepo      repository.StockRepo
	priceCache     *priceCache
	StockClient    external.StockClient
	TelegramClient external.TelegramClient
	PersonalChatID int64
	guard          runGuard
}

type DashboardPerformance struct {
	Delta      float64 `json:"delta"`
	Percentage float64 `json:"percentage"`
}

type DashboardSummary struct {
	JKSE      DashboardPerformance            `json:"jkse"`
	Portfolio DashboardPerformance            `json:"portfolio"`
	Positions map[string]DashboardPerformance `json:"positions"`
}

// DashboardProgressPoint compares each series' return from the first point in
// the selected range, so the Index and Portfolio remain comparable on one axis.
type DashboardProgressPoint struct {
	Date      string  `json:"date"`
	Index     float64 `json:"index"`
	Portfolio float64 `json:"portfolio"`
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

		if stock.Status == false && *stock.CurrentPrice <= stock.BestPrice {
			result = append(result, fmt.Sprintf("%s hitting best price", stock.Name))
		}

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

func (s *StockServiceImpl) InitializeSnapshots() {
	stocks, err := s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot initialize price snapshots: %v", err)
		return
	}
	s.cacheHistory(stocks)
}

func (s *StockServiceImpl) fetchAndUpdatePrices(stocks []DashboardStock) {
	s.guard.run("stock refresh", func() {
		if s.priceCache == nil {
			s.priceCache = newPriceCache()
		}
		for _, stock := range stocks {
			history, err := s.StockClient.GetStockHistory(stock.Name)
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
				continue
			}
			if len(history) == 0 {
				log.Printf("[ERROR] no price history for %s\n", stock.Name)
				continue
			}
			s.priceCache.put(stock.Name, history)
			currentPrice := int64(math.Round(history[len(history)-1].Close))

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
		history, err := s.StockClient.GetJKSEHistory()
		if err != nil {
			log.Printf("[ERROR] caching JKSE history: %v", err)
			return
		}
		s.priceCache.put("^JKSE", history)
	})
}

func (s *StockServiceImpl) RefreshPrices() ([]DashboardStock, error) {
	stocks, err := s.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve data from DB: %v\n", err)
		return nil, err
	}

	s.fetchAndUpdatePrices(stocks)
	updated, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *StockServiceImpl) GetAll() ([]DashboardStock, error) {
	stocks, err := s.StockRepo.GetAll()
	if err != nil {
		log.Printf("[ERROR] cannot retrieve stocks: %v\n", err)
		return nil, err
	}
	var dashboardStocks []DashboardStock
	for _, st := range stocks {
		dashboardStocks = append(dashboardStocks, DashboardStock(st))
	}
	return dashboardStocks, nil
}

func (s *StockServiceImpl) GetJKSE() (JKSEQuote, error) {
	quote, err := s.StockClient.GetJKSE()
	if err != nil {
		return JKSEQuote{}, err
	}
	return JKSEQuote{CurrentPrice: quote.CurrentPrice, PreviousClose: quote.PreviousClose}, nil
}

func (s *StockServiceImpl) GetSummary(period string) (DashboardSummary, error) {
	if period != "1d" && period != "5d" && period != "1mo" && period != "3mo" && period != "6mo" && period != "1y" && period != "ytd" {
		return DashboardSummary{}, ValidationError{Message: "invalid period"}
	}
	jkse, err := s.performance("^JKSE", period)
	if err != nil {
		return DashboardSummary{}, err
	}
	stocks, err := s.StockRepo.GetAll()
	if err != nil {
		return DashboardSummary{}, err
	}
	var current, reference float64
	positions := map[string]DashboardPerformance{}
	for _, stock := range stocks {
		if !stock.Status || stock.Lot == nil {
			continue
		}
		latest, baseline, err := s.priceCache.prices(stock.Name, period)
		if err != nil {
			return DashboardSummary{}, err
		}
		shares := float64(*stock.Lot * 100)
		current += latest * shares
		reference += baseline * shares
		positions[stock.Name] = performanceFrom(latest*shares, baseline*shares)
	}
	if reference == 0 {
		return DashboardSummary{}, errors.New("portfolio has no price history")
	}
	return DashboardSummary{JKSE: jkse, Portfolio: performanceFrom(current, reference), Positions: positions}, nil
}

func (s *StockServiceImpl) GetProgression(period string) ([]DashboardProgressPoint, error) {
	if period == "all" || period == "1d" {
		period = "5d"
	}
	if period != "5d" && period != "1mo" && period != "3mo" && period != "6mo" && period != "1y" && period != "ytd" {
		return nil, ValidationError{Message: "invalid period"}
	}
	if s.priceCache == nil {
		return nil, errors.New("price history is not available yet")
	}
	indexHistory := historyForPeriod(s.priceCache.get("^JKSE"), period)
	if len(indexHistory) == 0 {
		return nil, errors.New("index price history is not available yet")
	}
	stocks, err := s.StockRepo.GetAll()
	if err != nil {
		return nil, err
	}
	histories := make(map[string][]external.HistoricalPrice)
	for _, stock := range stocks {
		if stock.Status && stock.Lot != nil {
			histories[stock.Name] = s.priceCache.get(stock.Name)
		}
	}
	if len(histories) == 0 {
		return nil, errors.New("portfolio has no holdings")
	}

	points := make([]DashboardProgressPoint, 0, len(indexHistory))
	for _, index := range indexHistory {
		portfolio, ok := portfolioValueAt(stocks, histories, index.Date)
		if !ok {
			continue
		}
		points = append(points, DashboardProgressPoint{Date: index.Date.Format("2006-01-02"), Index: index.Close, Portfolio: portfolio})
	}
	if len(points) < 2 {
		return nil, errors.New("not enough price history for progression")
	}
	baseIndex, basePortfolio := points[0].Index, points[0].Portfolio
	for i := range points {
		points[i].Index = (points[i].Index/baseIndex - 1) * 100
		points[i].Portfolio = (points[i].Portfolio/basePortfolio - 1) * 100
	}
	return aggregateProgression(points, period), nil
}

func historyForPeriod(history []external.HistoricalPrice, period string) []external.HistoricalPrice {
	if len(history) == 0 {
		return nil
	}
	if period == "5d" {
		start := len(history) - 6
		if start < 0 {
			start = 0
		}
		return history[start:]
	}
	now := time.Now()
	targets := map[string]time.Time{
		"1mo": now.AddDate(0, -1, 0), "3mo": now.AddDate(0, -3, 0),
		"6mo": now.AddDate(0, -6, 0), "1y": now.AddDate(-1, 0, 0),
		"ytd": time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()),
	}
	for i, point := range history {
		if !point.Date.Before(targets[period]) {
			return history[i:]
		}
	}
	return history
}

func portfolioValueAt(stocks []repository.Stock, histories map[string][]external.HistoricalPrice, at time.Time) (float64, bool) {
	var total float64
	for _, stock := range stocks {
		if !stock.Status || stock.Lot == nil {
			continue
		}
		history := histories[stock.Name]
		index := -1
		for i := len(history) - 1; i >= 0; i-- {
			if !history[i].Date.After(at) {
				index = i
				break
			}
		}
		if index < 0 {
			return 0, false
		}
		total += history[index].Close * float64(*stock.Lot*100)
	}
	return total, total > 0
}

func aggregateProgression(points []DashboardProgressPoint, period string) []DashboardProgressPoint {
	if period == "5d" || period == "1mo" {
		return points
	}
	grouped := make([]DashboardProgressPoint, 0, len(points))
	lastKey := ""
	for _, point := range points {
		date, _ := time.Parse("2006-01-02", point.Date)
		key := date.Format("2006-01")
		if period == "3mo" || period == "6mo" {
			year, week := date.ISOWeek()
			key = fmt.Sprintf("%d-%02d", year, week)
		}
		if key != lastKey {
			grouped = append(grouped, point)
			lastKey = key
			continue
		}
		grouped[len(grouped)-1] = point // use the last trading day in the bucket
	}
	return grouped
}

func (s *StockServiceImpl) performance(symbol, period string) (DashboardPerformance, error) {
	latest, reference, err := s.priceCache.prices(symbol, period)
	if err != nil {
		return DashboardPerformance{}, err
	}
	return performanceFrom(latest, reference), nil
}

func performanceFrom(current, reference float64) DashboardPerformance {
	delta := current - reference
	return DashboardPerformance{Delta: delta, Percentage: (delta / reference) * 100}
}

func (s *StockServiceImpl) cacheHistory(stocks []DashboardStock) {
	if s.priceCache == nil {
		s.priceCache = newPriceCache()
	}
	for _, stock := range stocks {
		history, err := s.StockClient.GetStockHistory(stock.Name)
		if err != nil {
			log.Printf("[ERROR] caching %s history: %v", stock.Name, err)
			continue
		}
		s.priceCache.put(stock.Name, history)
	}
	history, err := s.StockClient.GetJKSEHistory()
	if err != nil {
		log.Printf("[ERROR] caching JKSE history: %v", err)
		return
	}
	s.priceCache.put("^JKSE", history)
}

func (s *StockServiceImpl) Create(stock DashboardStock) (string, error) {
	if stock.BestPrice <= 0 || stock.FairPrice <= 0 {
		return "", ValidationError{Message: "best_price and fair_price are required and must be > 0"}
	}
	st := repository.Stock(stock)
	name, err := s.StockRepo.Create(st)
	if err != nil {
		log.Printf("[ERROR] cannot create stock: %v\n", err)
	}
	return name, err
}

func (s *StockServiceImpl) Update(stock DashboardStock) (string, error) {
	if stock.BestPrice <= 0 || stock.FairPrice <= 0 {
		return "", ValidationError{Message: "best_price and fair_price are required and must be > 0"}
	}
	st := repository.Stock(stock)
	name, err := s.StockRepo.Update(st)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Printf("[ERROR] cannot update stock: %v\n", err)
	}
	return name, err
}

func (s *StockServiceImpl) Delete(name string) (string, error) {
	deletedName, err := s.StockRepo.Delete(name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Printf("[ERROR] cannot delete stock: %v\n", err)
	}
	return deletedName, err
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
