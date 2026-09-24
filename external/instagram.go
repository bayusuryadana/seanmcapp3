package external

import (
	"fmt"
	"io"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

const (
	igAppID     = "936619743392459"
	igUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)

type InstagramClient interface {
	Get(url string) ([]byte, error)
}

type InstagramClientImpl struct {
	Client tls_client.HttpClient
}

func (c *InstagramClientImpl) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("Referer", "https://www.instagram.com/")
	req.Header.Set("User-Agent", igUserAgent)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
