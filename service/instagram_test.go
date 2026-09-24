package service

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"seanmcapp/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMain(m *testing.M) {
	sleepFn = func(time.Duration) {}
	os.Exit(m.Run())
}

func TestDetectNewPosts(t *testing.T) {
	posts := []igPost{{Shortcode: "A"}, {Shortcode: "B"}, {Shortcode: "C"}}
	assert.Nil(t, detectNewPosts("", posts))
	assert.Equal(t, []igPost{{Shortcode: "C"}}, detectNewPosts("A, B", posts))
}

func TestProcessAccountFetchesPostsOnly(t *testing.T) {
	var requests []string
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		requests = append(requests, url)
		return []byte(`{"items":[{"code":"ABC","image_versions2":{"candidates":[{"url":"https://image"}]}}]}`), nil
	}}
	repo := &fakeInstagramRepo{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: repo, InstagramClient: client, TelegramClient: &fakeTelegramClient{}}
	svc.processAccount(repository.InstagramAccount{Username: "foo", UserID: "123", LastShortcodes: "ABC"})
	require.Len(t, requests, 1)
	assert.Contains(t, requests[0], igFeedBase+"123")
	assert.Equal(t, "ABC", repo.updatedShortcodes["foo"])
}

func TestProcessAccountHandlesPostFetchFailure(t *testing.T) {
	svc := &InstagramServiceImpl{InstagramAccountRepo: &fakeInstagramRepo{}, InstagramClient: &fakeInstagramClient{getFn: func(string) ([]byte, error) { return nil, errors.New("failed") }}}
	svc.processAccount(repository.InstagramAccount{Username: "foo", UserID: "123"})
}

func TestResolveUserID(t *testing.T) {
	t.Run("uses stored id", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{getFn: func(string) ([]byte, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}}}
		id, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo", UserID: "123"})
		require.NoError(t, err)
		assert.Equal(t, "123", id)
	})

	t.Run("fetches and persists id", func(t *testing.T) {
		repo := &fakeInstagramRepo{}
		svc := &InstagramServiceImpl{InstagramAccountRepo: repo, InstagramClient: &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
			assert.Equal(t, igProfileBase+"foo", url)
			return []byte(`{"data":{"user":{"id":"123"}}}`), nil
		}}}
		id, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
		require.NoError(t, err)
		assert.Equal(t, "123", id)
		assert.Equal(t, "123", repo.updatedUserIDs["foo"])
	})

	for _, tc := range []struct {
		name string
		body []byte
		err  error
	}{
		{name: "request failure", err: errors.New("network")},
		{name: "invalid profile", body: []byte(`{"data":{"user":{}}}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{getFn: func(string) ([]byte, error) { return tc.body, tc.err }}}
			_, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
			assert.Error(t, err)
		})
	}
}

func TestFetchLatestPostsAndMediaExtraction(t *testing.T) {
	feed := `{"items":[
		{"code":"PHOTO","caption":{"text":"hello"},"image_versions2":{"candidates":[{"url":"https://photo"}]}},
		{"code":"CAROUSEL","media_type":8,"carousel_media":[
			{"image_versions2":{"candidates":[{"url":"https://carousel-photo"}]}},
			{"media_type":2,"video_versions":[{"url":"https://video"}],"image_versions2":{"candidates":[{"url":"https://thumb"}]}}
		]},
		{"code":"NO_MEDIA"}, {"image_versions2":{"candidates":[{"url":"https://missing-code"}]}}
	]}`
	svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		assert.Equal(t, fmt.Sprintf("%s123/?count=%d", igFeedBase, igMaxPosts), url)
		return []byte(feed), nil
	}}}
	posts, err := svc.fetchLatestPosts("foo", "123")
	require.NoError(t, err)
	require.Len(t, posts, 2)
	assert.Equal(t, "hello", posts[0].Caption)
	assert.Equal(t, "https://photo", posts[0].Media[0].URL)
	require.Len(t, posts[1].Media, 2)
	assert.False(t, posts[1].Media[0].IsVideo)
	assert.True(t, posts[1].Media[1].IsVideo)
	assert.Equal(t, "https://thumb", posts[1].Media[1].ThumbnailURL)

	assert.Nil(t, extractMedia(gjsonParse(`{"media_type":2}`)))
	assert.Nil(t, extractMedia(gjsonParse(`{"media_type":8,"carousel_media":[]}`)))
}

func TestFetchLatestPostsErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []byte
		err  error
	}{
		{name: "request error", err: errors.New("network")},
		{name: "unexpected response", body: []byte(`{}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{getFn: func(string) ([]byte, error) { return tc.body, tc.err }}}
			_, err := svc.fetchLatestPosts("foo", "123")
			assert.Error(t, err)
		})
	}
}

func TestSchedulingAndRun(t *testing.T) {
	accounts := []repository.InstagramAccount{{ID: 1, Username: "one"}, {ID: 2, Username: "two"}}
	assert.Empty(t, selectAccountsForHour(nil, 0))
	assert.Equal(t, []repository.InstagramAccount{accounts[0]}, selectAccountsForHour(accounts, 0))
	assert.Equal(t, 100*time.Millisecond, randomDuration(100*time.Millisecond, 100*time.Millisecond))

	previousHourFn := hourFn
	hourFn = func() int { return 0 }
	t.Cleanup(func() { hourFn = previousHourFn })
	repo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) { return accounts[:1], nil }}
	svc := &InstagramServiceImpl{InstagramAccountRepo: repo, InstagramClient: &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return []byte(`{"items":[]}`), nil
	}}, TelegramClient: &fakeTelegramClient{}}
	svc.Run()
	assert.Equal(t, "", repo.updatedShortcodes["one"])

	(&InstagramServiceImpl{InstagramAccountRepo: &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return nil, errors.New("database")
	}}}).Run()
}

func TestPostNotificationAndVideoFallbacks(t *testing.T) {
	t.Run("sends photo and summary", func(t *testing.T) {
		tg := &fakeTelegramClient{}
		svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}
		svc.notify("foo_bar", []igPost{{Shortcode: "ABC", Caption: "a_b", Media: []igMedia{{URL: "https://photo"}}}})
		require.Len(t, tg.photos, 1)
		require.Len(t, tg.messages, 1)
		assert.Contains(t, tg.messages[0].text, "foo\\_bar")
		assert.Contains(t, tg.messages[0].text, "a\\_b")
	})

	t.Run("keeps directly accepted video", func(t *testing.T) {
		tg := &fakeTelegramClient{}
		svc := &InstagramServiceImpl{TelegramClient: tg}
		svc.sendMedia("foo", "ABC", "https://post", 0, igMedia{IsVideo: true, URL: "https://video"})
		assert.Len(t, tg.videos, 1)
		assert.Empty(t, tg.uploads)
	})

	for _, tc := range []struct {
		name         string
		get          func(string) ([]byte, error)
		thumbnail    string
		uploadFails  bool
		wantMessages int
		wantPhotos   int
		wantUploads  int
	}{
		{name: "download failure without thumbnail", get: func(string) ([]byte, error) { return nil, errors.New("download") }, wantMessages: 1},
		{name: "oversized download uses thumbnail", get: func(string) ([]byte, error) { return make([]byte, igMaxUploadBytes+1), nil }, thumbnail: "https://thumb", wantPhotos: 1},
		{name: "upload failure uses thumbnail", get: func(string) ([]byte, error) { return []byte("video"), nil }, thumbnail: "https://thumb", uploadFails: true, wantUploads: 1, wantPhotos: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tg := &fakeTelegramClient{videoURLFails: true, uploadFails: tc.uploadFails}
			svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: &fakeInstagramClient{getFn: tc.get}}
			svc.sendMedia("foo", "ABC", "https://post", 2, igMedia{IsVideo: true, URL: "https://video", ThumbnailURL: tc.thumbnail})
			assert.Len(t, tg.messages, tc.wantMessages)
			assert.Len(t, tg.photos, tc.wantPhotos)
			assert.Len(t, tg.uploads, tc.wantUploads)
		})
	}

	assert.Equal(t, "\\_\\*\\`\\[", escapeMarkdown("_*`["))
}

func gjsonParse(raw string) gjson.Result { return gjson.Parse(raw) }
