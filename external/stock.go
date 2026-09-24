package external

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type StockClient interface {
	GetPrice(name string) (int64, error)
	GetJKSE() (IndexQuote, error)
	GetStockHistory(name string) ([]HistoricalPrice, error)
	GetJKSEHistory() ([]HistoricalPrice, error)
}

type IndexQuote struct {
	CurrentPrice  float64
	PreviousClose float64
}

type HistoricalPrice struct {
	Date  time.Time
	Close float64
}

type StockClientImpl struct {
	Client *http.Client
}

var stockURLTemplate = "https://query1.finance.yahoo.com/v8/finance/chart/{{name}}.jk"
var jkseURL = "https://query1.finance.yahoo.com/v8/finance/chart/%5EJKSE"
var historyURLTemplate = "https://query1.finance.yahoo.com/v8/finance/chart/{{symbol}}?range=1y&interval=1d"

const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func (s *StockClientImpl) GetPrice(name string) (int64, error) {
	stockURL := strings.NewReplacer("{{name}}", name).Replace(stockURLTemplate)
	meta, err := s.getMeta(stockURL)
	if err != nil {
		return 0, err
	}

	regularMarketPrice := gjson.Get(meta, "regularMarketPrice")
	if !regularMarketPrice.Exists() {
		return 0, fmt.Errorf("stock %s not found in json", name)
	}

	return regularMarketPrice.Int(), nil
}

func (s *StockClientImpl) GetJKSE() (IndexQuote, error) {
	meta, err := s.getMeta(jkseURL)
	if err != nil {
		return IndexQuote{}, err
	}

	currentPrice := gjson.Get(meta, "regularMarketPrice")
	previousClose := gjson.Get(meta, "previousClose")
	if !previousClose.Exists() {
		previousClose = gjson.Get(meta, "regularMarketPreviousClose")
	}
	if !currentPrice.Exists() {
		return IndexQuote{}, fmt.Errorf("JKSE quote not found in json")
	}

	quote := IndexQuote{CurrentPrice: currentPrice.Float()}
	if previousClose.Exists() {
		quote.PreviousClose = previousClose.Float()
	}
	return quote, nil
}

func (s *StockClientImpl) GetStockHistory(name string) ([]HistoricalPrice, error) {
	url := strings.NewReplacer("{{symbol}}", name+".jk").Replace(historyURLTemplate)
	return s.getHistory(url)
}

func (s *StockClientImpl) GetJKSEHistory() ([]HistoricalPrice, error) {
	url := strings.NewReplacer("{{symbol}}", "%5EJKSE").Replace(historyURLTemplate)
	return s.getHistory(url)
}

func (s *StockClientImpl) getMeta(stockURL string) (string, error) {

	req, err := http.NewRequest(http.MethodGet, stockURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot build request: %w", err)
	}
	req.Header.Set("User-Agent", browserUserAgent)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot fetch stock data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	meta := gjson.Get(string(body), "chart.result.0.meta")
	if !meta.Exists() {
		return "", fmt.Errorf("stock quote not found in json")
	}
	return meta.Raw, nil
}

func (s *StockClientImpl) getHistory(stockURL string) ([]HistoricalPrice, error) {
	req, err := http.NewRequest(http.MethodGet, stockURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot build request: %w", err)
	}
	req.Header.Set("User-Agent", browserUserAgent)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch stock data: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	timestamps := gjson.GetBytes(body, "chart.result.0.timestamp").Array()
	closes := gjson.GetBytes(body, "chart.result.0.indicators.quote.0.close").Array()
	if len(timestamps) == 0 || len(timestamps) != len(closes) {
		return nil, fmt.Errorf("stock history not found in json")
	}

	history := make([]HistoricalPrice, 0, len(closes))
	for i, close := range closes {
		if !close.Exists() || close.Type == gjson.Null {
			continue
		}
		history = append(history, HistoricalPrice{Date: time.Unix(timestamps[i].Int(), 0), Close: close.Float()})
	}
	return history, nil
}
