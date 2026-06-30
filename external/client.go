package external

import (
	"net/http"
	"time"
)

const httpTimeout = 15 * time.Second

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}
