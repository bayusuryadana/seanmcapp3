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

func TestStockGetPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":1234}}]}}`))
	}))
	defer srv.Close()
	defer withStockURL(srv.URL + "/{{name}}")()

	price, err := NewStockClient().GetPrice("BBCA")
	require.NoError(t, err)
	assert.Equal(t, int64(1234), price)
}

func TestStockGetPriceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[]}}`))
	}))
	defer srv.Close()
	defer withStockURL(srv.URL + "/{{name}}")()

	_, err := NewStockClient().GetPrice("BBCA")
	assert.Error(t, err)
}

func TestStockGetPriceRequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	defer withStockURL(url + "/{{name}}")()

	_, err := NewStockClient().GetPrice("BBCA")
	assert.Error(t, err)
}
