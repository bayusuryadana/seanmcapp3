package external

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func instagramTestClient(t *testing.T) *InstagramClientImpl {
	t.Helper()
	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithTimeoutSeconds(15),
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
	)
	require.NoError(t, err)
	return &InstagramClientImpl{Client: client}
}

func TestInstagramGetOK(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			assert.Contains(t, r.Header.Get("Cookie"), "sessionid=sid")
			assert.Equal(t, "csrf", r.Header.Get("X-CSRFToken"))
		} else {
			assert.Contains(t, r.Header.Get("Cookie"), "sessionid=new-sid")
			assert.Equal(t, "new-csrf", r.Header.Get("X-CSRFToken"))
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := instagramTestClient(t)
	client.SetCredentials("sid", "csrf")
	body, err := client.Get(srv.URL)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))

	client.SetCredentials("new-sid", "new-csrf")
	_, err = client.Get(srv.URL)
	require.NoError(t, err)
}

func TestInstagramGetUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := instagramTestClient(t)
	client.SetCredentials("sid", "csrf")
	_, err := client.Get(srv.URL)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionExpired)
}

func TestInstagramGetUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := instagramTestClient(t)
	client.SetCredentials("sid", "csrf")
	_, err := client.Get(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}
