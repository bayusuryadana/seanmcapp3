package external

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstagramGetOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Required auth headers are attached.
		assert.Contains(t, r.Header.Get("Cookie"), "sessionid=sid")
		assert.Equal(t, "csrf", r.Header.Get("X-CSRFToken"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, err := NewInstagramClient("sid", "csrf").Get(srv.URL)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestInstagramGetUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := NewInstagramClient("sid", "csrf").Get(srv.URL)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionExpired)
}

func TestInstagramGetUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewInstagramClient("sid", "csrf").Get(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}
