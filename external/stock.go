package external

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type StockClient interface {
	GetPrice(name string) (int64, error)
}

type StockClientImpl struct {
	client *http.Client
}

func NewStockClient() *StockClientImpl {
	return &StockClientImpl{client: newHTTPClient()}
}

var stockURLTemplate = "https://query1.finance.yahoo.com/v8/finance/chart/{{name}}.jk"

const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func (s *StockClientImpl) GetPrice(name string) (int64, error) {
	stockURL := strings.NewReplacer("{{name}}", name).Replace(stockURLTemplate)

	req, err := http.NewRequest(http.MethodGet, stockURL, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot build request: %w", err)
	}
	req.Header.Set("User-Agent", browserUserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot fetch stock data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("reading response: %w", err)
	}

	regularMarketPrice := gjson.Get(string(body), "chart.result.0.meta.regularMarketPrice")
	if !regularMarketPrice.Exists() {
		return 0, fmt.Errorf("stock %s not found in json", name)
	}

	return regularMarketPrice.Int(), nil
}

