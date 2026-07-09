package external

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
)

const (
	igAppID     = "936619743392459"
	igUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)

// ErrSessionExpired is returned when Instagram rejects the request with 401/403,
// which usually means the configured IG_SESSION_ID is no longer valid.
var ErrSessionExpired = errors.New("instagram session expired or blocked — please update IG_SESSION_ID")

type InstagramClient interface {
	Get(url string) ([]byte, error)
}

type InstagramClientImpl struct {
	SessionID string
	CSRFToken string
	client    *http.Client
}

func NewInstagramClient(sessionID, csrfToken string) *InstagramClientImpl {
	jar, _ := cookiejar.New(nil)
	return &InstagramClientImpl{
		SessionID: sessionID,
		CSRFToken: csrfToken,
		client:    &http.Client{Timeout: httpTimeout, Jar: jar},
	}
}

func (c *InstagramClientImpl) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", fmt.Sprintf("sessionid=%s; csrftoken=%s", c.SessionID, c.CSRFToken))
	req.Header.Set("X-CSRFToken", c.CSRFToken)
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("Referer", "https://www.instagram.com/")
	req.Header.Set("User-Agent", igUserAgent)

	resp, err := c.client.Do(req)
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
