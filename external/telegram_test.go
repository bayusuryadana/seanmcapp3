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

func TestTelegramSendPhotoErrors(t *testing.T) {
	t.Run("decode error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not-json`))
		}))
		defer srv.Close()
		_, err := NewTelegramClient(srv.URL, "bot").SendPhoto(5, "http://img", "cap")
		assert.Error(t, err)
	})

	t.Run("request error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()
		_, err := NewTelegramClient(url, "bot").SendPhoto(5, "http://img", "cap")
		assert.Error(t, err)
	})
}

func TestTelegramSendVideo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sendvideo")
		assert.Equal(t, "http://vid/1", r.URL.Query().Get("video"))
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":12,"chat":{"id":5,"type":"private"}}}`))
	}))
	defer srv.Close()

	resp, err := NewTelegramClient(srv.URL, "bot").SendVideo(5, "http://vid/1", "cap")
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, 12, resp.Result.MessageID)
}

func TestTelegramSendVideoUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/sendvideo")
		require.NoError(t, r.ParseMultipartForm(1<<20))
		assert.Equal(t, "5", r.FormValue("chat_id"))
		file, hdr, err := r.FormFile("video")
		require.NoError(t, err)
		defer file.Close()
		assert.Equal(t, "clip.mp4", hdr.Filename)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":13,"chat":{"id":5,"type":"private"}}}`))
	}))
	defer srv.Close()

	resp, err := NewTelegramClient(srv.URL, "bot").SendVideoUpload(5, []byte("bytes"), "clip.mp4", "cap")
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, 13, resp.Result.MessageID)
}

