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
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), tls_client.WithClientProfile(profiles.Chrome_144))
	require.NoError(t, err)
	return &InstagramClientImpl{Client: client}
}

func TestInstagramGetIsUnauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Cookie"))
		assert.Empty(t, r.Header.Get("X-CSRFToken"))
		assert.NotEmpty(t, r.Header.Get("X-IG-App-ID"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, err := instagramTestClient(t).Get(srv.URL)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestInstagramGetReportsUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer srv.Close()
	_, err := instagramTestClient(t).Get(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}
