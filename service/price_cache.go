package service

import (
	"errors"
	"seanmcapp/external"
	"sync"
	"time"
)

type priceCache struct {
	mu      sync.RWMutex
	history map[string][]external.HistoricalPrice
}

func newPriceCache() *priceCache {
	return &priceCache{history: map[string][]external.HistoricalPrice{}}
}

func (c *priceCache) put(symbol string, history []external.HistoricalPrice) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history[symbol] = history
}

func (c *priceCache) get(symbol string) []external.HistoricalPrice {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]external.HistoricalPrice(nil), c.history[symbol]...)
}

func (c *priceCache) prices(symbol, period string) (float64, float64, error) {
	c.mu.RLock()
	history := c.history[symbol]
	c.mu.RUnlock()
	if len(history) < 2 {
		return 0, 0, errors.New("price history is not available yet")
	}
	latest := history[len(history)-1].Close
	var index int
	switch period {
	case "1d":
		index = len(history) - 2
	case "5d":
		index = len(history) - 6
	case "1mo", "3mo", "6mo", "1y", "ytd":
		months := map[string]int{"1mo": 1, "3mo": 3, "6mo": 6, "1y": 12}
		target := time.Now().AddDate(0, -months[period], 0)
		if period == "ytd" {
			now := time.Now()
			target = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		}
		index = -1
		for i := len(history) - 1; i >= 0; i-- {
			if !history[i].Date.After(target) {
				index = i
				break
			}
		}
	default:
		return 0, 0, errors.New("invalid period")
	}
	if index < 0 {
		return 0, 0, errors.New("not enough price history")
	}
	return latest, history[index].Close, nil
}
