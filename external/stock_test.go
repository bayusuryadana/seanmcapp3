package external

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withStockURL(url string) func() {
	original := stockURLTemplate
	stockURLTemplate = url
	return func() { stockURLTemplate = original }
}

func withJKSEURL(url string) func() {
	original := jkseURL
	jkseURL = url
	return func() { jkseURL = original }
}

func withHistoryURL(url string) func() {
	original := historyURLTemplate
	historyURLTemplate = url
	return func() { historyURLTemplate = original }
}

func TestStockGetPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":1234}}]}}`))
	}))
	defer srv.Close()
	defer withStockURL(srv.URL + "/{{name}}")()

	price, err := (&StockClientImpl{Client: srv.Client()}).GetPrice("BBCA")
	require.NoError(t, err)
	assert.Equal(t, int64(1234), price)
}

func TestStockGetPriceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[]}}`))
	}))
	defer srv.Close()
	defer withStockURL(srv.URL + "/{{name}}")()

	_, err := (&StockClientImpl{Client: srv.Client()}).GetPrice("BBCA")
	assert.Error(t, err)
}

func TestStockGetPriceRequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	defer withStockURL(url + "/{{name}}")()

	_, err := (&StockClientImpl{Client: srv.Client()}).GetPrice("BBCA")
	assert.Error(t, err)
}

func TestStockGetJKSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/%5EJKSE", r.URL.EscapedPath())
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":7123.45,"previousClose":7000.00}}]}}`))
	}))
	defer srv.Close()
	defer withJKSEURL(srv.URL + "/%5EJKSE")()

	quote, err := (&StockClientImpl{Client: srv.Client()}).GetJKSE()
	require.NoError(t, err)
	assert.Equal(t, 7123.45, quote.CurrentPrice)
	assert.Equal(t, 7000.0, quote.PreviousClose)
}

func TestStockHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "interval=1d")
		_, _ = w.Write([]byte(`{"chart":{"result":[{"timestamp":[1704067200,1704153600],"indicators":{"quote":[{"close":[100.5,101.25]}]}}]}}`))
	}))
	defer srv.Close()
	defer withHistoryURL(srv.URL + "/{{symbol}}?range=1y&interval=1d")()

	client := &StockClientImpl{Client: srv.Client()}
	stock, err := client.GetStockHistory("BBCA")
	require.NoError(t, err)
	require.Len(t, stock, 2)
	assert.Equal(t, 101.25, stock[1].Close)

	index, err := client.GetJKSEHistory()
	require.NoError(t, err)
	assert.Len(t, index, 2)
}
