package external

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramSendMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sendmessage")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":10,"chat":{"id":5,"type":"private"}}}`))
	}))
	defer srv.Close()

	c := NewTelegramClient(srv.URL, "bot")
	resp, err := c.SendMessage(5, "hello")
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, 10, resp.Result.MessageID)
}

func TestTelegramSendPhoto(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sendphoto")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":11,"chat":{"id":5,"type":"private"}}}`))
	}))
	defer srv.Close()

	c := NewTelegramClient(srv.URL, "bot")
	resp, err := c.SendPhoto(5, "http://img/1", "caption")
	require.NoError(t, err)
	assert.True(t, resp.Ok)
}

func TestTelegramDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	c := NewTelegramClient(srv.URL, "bot")
	_, err := c.SendMessage(5, "hello")
	assert.Error(t, err)
}

func TestTelegramRequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // server no longer listening -> request fails

	c := NewTelegramClient(url, "bot")
	_, err := c.SendMessage(5, "hello")
	assert.Error(t, err)
}
