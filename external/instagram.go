package external

import (
	"errors"
	"fmt"
	"io"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

const (
	igAppID     = "936619743392459"
	igUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)

var ErrSessionExpired = errors.New("instagram session expired or blocked")

type InstagramClient interface {
	SetCredentials(sessionID, csrfToken string)
	Get(url string) ([]byte, error)
}

type InstagramClientImpl struct {
	sessionID string
	csrfToken string
	Client    tls_client.HttpClient
	mu        sync.RWMutex
}

func (c *InstagramClientImpl) SetCredentials(sessionID, csrfToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = sessionID
	c.csrfToken = csrfToken
}

func (c *InstagramClientImpl) Get(url string) ([]byte, error) {
	c.mu.RLock()
	sessionID := c.sessionID
	csrfToken := c.csrfToken
	c.mu.RUnlock()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", fmt.Sprintf("sessionid=%s; csrftoken=%s", sessionID, csrfToken))
	req.Header.Set("X-CSRFToken", csrfToken)
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("Referer", "https://www.instagram.com/")
	req.Header.Set("User-Agent", igUserAgent)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%w (HTTP %d)", ErrSessionExpired, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
