package service

import (
	"errors"
	"fmt"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectNewPosts(t *testing.T) {
	tests := []struct {
		name    string
		stored  string
		current []igPost
		want    []string // shortcodes expected as "new"
	}{
		{"no stored history returns nothing", "", []igPost{{Shortcode: "A"}}, nil},
		{"one new post", "A,B", []igPost{{Shortcode: "A"}, {Shortcode: "B"}, {Shortcode: "C"}}, []string{"C"}},
		{"no new posts", "A,B", []igPost{{Shortcode: "A"}, {Shortcode: "B"}}, nil},
		{"handles whitespace in stored", "A, B", []igPost{{Shortcode: "C"}}, []string{"C"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectNewPosts(tc.stored, tc.current)
			var codes []string
			for _, p := range got {
				codes = append(codes, p.Shortcode)
			}
			assert.Equal(t, tc.want, codes)
		})
	}
}

const igProfileJSON = `{"data":{"user":{"id":"123"}}}`
const igFeedJSON = `{"items":[
	{"code":"AAA","image_versions2":{"candidates":[{"url":"http://img/a"}]}},
	{"code":"BBB","image_versions2":{"candidates":[{"url":"http://img/b"}]}}
]}`

func TestFetchLatestPosts(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, "web_profile_info") {
			return []byte(igProfileJSON), nil
		}
		return []byte(igFeedJSON), nil
	}}
	repo := &fakeInstagramRepo{}
	svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: repo}

	posts, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo"})
	require.NoError(t, err)
	require.Len(t, posts, 2)
	assert.Equal(t, "AAA", posts[0].Shortcode)
	require.Len(t, posts[0].Media, 1)
	assert.False(t, posts[0].Media[0].IsVideo)
	assert.Equal(t, "http://img/a", posts[0].Media[0].URL)
	// user id was resolved and persisted
	assert.Equal(t, "123", repo.updatedUserIDs["foo"])
}

func TestFetchLatestPostsSkipsProfileWhenUserIDKnown(t *testing.T) {
	var calledURLs []string
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		calledURLs = append(calledURLs, url)
		return []byte(igFeedJSON), nil
	}}
	repo := &fakeInstagramRepo{}
	svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: repo}

	posts, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo", UserID: "123"})
	require.NoError(t, err)
	require.Len(t, posts, 2)
	// profile endpoint must not be called, and user id must not be re-persisted
	for _, u := range calledURLs {
		assert.NotContains(t, u, "web_profile_info")
	}
	assert.Empty(t, repo.updatedUserIDs)
}

func TestFetchLatestPostsErrors(t *testing.T) {
	t.Run("profile request error", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return nil, errors.New("net") },
		}}
		_, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo"})
		assert.Error(t, err)
	})

	t.Run("empty user id", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return []byte(`{"data":{"user":{}}}`), nil },
		}}
		_, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo"})
		assert.Error(t, err)
	})

	t.Run("missing items in feed", func(t *testing.T) {
		svc := &InstagramServiceImpl{
			InstagramAccountRepo: &fakeInstagramRepo{},
			InstagramClient: &fakeInstagramClient{
				getFn: func(url string) ([]byte, error) {
					if strings.Contains(url, "web_profile_info") {
						return []byte(igProfileJSON), nil
					}
					return []byte(`{}`), nil
				},
			},
		}
		_, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo"})
		assert.Error(t, err)
	})
}

func TestInstagramRun(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{{Username: "foo", LastShortcodes: "AAA,BBB"}}, nil
	}}
	// Feed returns exactly the stored shortcodes -> no new posts -> no photo, but shortcodes are persisted.
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, "web_profile_info") {
			return []byte(igProfileJSON), nil
		}
		return []byte(igFeedJSON), nil
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg}

	svc.Run()

	assert.Empty(t, tg.photos)
	assert.Equal(t, "AAA,BBB", accountRepo.updatedShortcodes["foo"])
}

func TestInstagramRunSendsNewPosts(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{{Username: "foo", LastShortcodes: "AAA"}}, nil
	}}
	// Feed returns AAA + BBB, only AAA is known -> BBB is a new post -> notify.
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, "web_profile_info") {
			return []byte(igProfileJSON), nil
		}
		return []byte(igFeedJSON), nil
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg, PersonalChatID: 42}

	svc.Run()

	require.Len(t, tg.photos, 1)
	assert.Equal(t, int64(42), tg.photos[0].chatID)
	assert.Equal(t, "http://img/b", tg.photos[0].url)
	assert.Empty(t, tg.photos[0].caption) // media carries no caption anymore
	// the header + link now arrive as a separate summary message
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "foo")
	assert.Contains(t, tg.messages[0].text, "BBB")
	assert.Equal(t, "AAA,BBB", accountRepo.updatedShortcodes["foo"])
}

func TestInstagramRunAlertsOnceOnExpiredSession(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{
			{Username: "foo", UserID: "1"},
			{Username: "bar", UserID: "2"},
		}, nil
	}}
	// Feed endpoint always reports an expired session.
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return nil, fmt.Errorf("%w (HTTP 401)", external.ErrSessionExpired)
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg, PersonalChatID: 42}

	svc.Run()

	// Exactly one alert, and the run stops before touching the second account.
	require.Len(t, tg.messages, 1)
	assert.Equal(t, int64(42), tg.messages[0].chatID)
	assert.Contains(t, tg.messages[0].text, "IG_SESSION_ID")
	assert.Empty(t, tg.photos)
	assert.Empty(t, accountRepo.updatedShortcodes)
}

const igMixedFeedJSON = `{"items":[
	{"code":"VID","media_type":2,"caption":{"text":"a *fancy* caption"},"video_versions":[{"url":"http://vid/v"}],"image_versions2":{"candidates":[{"url":"http://img/vthumb"}]}},
	{"code":"CAR","media_type":8,"carousel_media":[
		{"media_type":1,"image_versions2":{"candidates":[{"url":"http://img/c1"}]}},
		{"media_type":2,"video_versions":[{"url":"http://vid/c2"}],"image_versions2":{"candidates":[{"url":"http://img/c2thumb"}]}}
	]}
]}`

func TestParseVideoAndCarousel(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte(igMixedFeedJSON), nil
	}}
	svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: &fakeInstagramRepo{}}

	posts, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo", UserID: "123"})
	require.NoError(t, err)
	require.Len(t, posts, 2)

	// video post
	require.Len(t, posts[0].Media, 1)
	assert.True(t, posts[0].Media[0].IsVideo)
	assert.Equal(t, "http://vid/v", posts[0].Media[0].URL)
	assert.Equal(t, "http://img/vthumb", posts[0].Media[0].ThumbnailURL)
	assert.Equal(t, "a *fancy* caption", posts[0].Caption)

	// carousel with an image + a video child, in order
	require.Len(t, posts[1].Media, 2)
	assert.False(t, posts[1].Media[0].IsVideo)
	assert.Equal(t, "http://img/c1", posts[1].Media[0].URL)
	assert.True(t, posts[1].Media[1].IsVideo)
	assert.Equal(t, "http://vid/c2", posts[1].Media[1].URL)
}

func TestNotifySendsVideoByURLThenSummary(t *testing.T) {
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}
	post := igPost{
		Shortcode: "VID",
		Caption:   "hello _world_",
		Media:     []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}},
	}

	svc.notify("foo", []igPost{post})

	require.Len(t, tg.videos, 1)
	assert.Equal(t, "http://vid/v", tg.videos[0].url)
	assert.Empty(t, tg.uploads) // URL send succeeded, no upload needed
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "foo")
	assert.Contains(t, tg.messages[0].text, "\\_world\\_") // caption markdown-escaped
}

func TestNotifyVideoUploadFallbackWhenURLFails(t *testing.T) {
	tg := &fakeTelegramClient{videoURLFails: true}
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte("small-video-bytes"), nil
	}}
	svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: client, PersonalChatID: 7}
	post := igPost{Shortcode: "VID", Media: []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}}}

	svc.notify("foo", []igPost{post})

	require.Len(t, tg.videos, 1)  // URL attempt happened
	require.Len(t, tg.uploads, 1) // fell back to multipart upload
	assert.Equal(t, "VID_0.mp4", tg.uploads[0].filename)
	assert.Empty(t, tg.photos) // upload succeeded, no thumbnail fallback
}

func TestNotifyVideoThumbnailFallbackWhenTooLarge(t *testing.T) {
	tg := &fakeTelegramClient{videoURLFails: true}
	big := make([]byte, igMaxUploadBytes+1)
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return big, nil
	}}
	svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: client, PersonalChatID: 7}
	post := igPost{Shortcode: "VID", Media: []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}}}

	svc.notify("foo", []igPost{post})

	assert.Empty(t, tg.uploads) // too big to upload
	require.Len(t, tg.photos, 1)
	assert.Equal(t, "http://img/t", tg.photos[0].url)
	assert.Contains(t, tg.photos[0].caption, "Instagram")
}

func TestNotifyVideoDownloadErrorFallsBack(t *testing.T) {
	tg := &fakeTelegramClient{videoURLFails: true}
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return nil, errors.New("download failed")
	}}
	svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: client, PersonalChatID: 7}
	post := igPost{Shortcode: "VID", Media: []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}}}

	svc.notify("foo", []igPost{post})

	assert.Empty(t, tg.uploads)
	require.Len(t, tg.photos, 1)
	assert.Equal(t, "http://img/t", tg.photos[0].url)
	assert.Contains(t, tg.photos[0].caption, "Instagram")
}

func TestNotifyVideoUploadFailureFallsBack(t *testing.T) {
	tg := &fakeTelegramClient{videoURLFails: true, uploadFails: true}
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte("small"), nil
	}}
	svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: client, PersonalChatID: 7}
	post := igPost{Shortcode: "VID", Media: []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}}}

	svc.notify("foo", []igPost{post})

	require.Len(t, tg.uploads, 1) // upload attempted
	require.Len(t, tg.photos, 1)  // then fell back to thumbnail
	assert.Equal(t, "http://img/t", tg.photos[0].url)
}

func TestSendVideoFallbackNoThumbnailSendsNote(t *testing.T) {
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}

	svc.sendVideoFallback("foo", "VID", "http://link/VID/", igMedia{IsVideo: true, URL: "http://vid/v"})

	assert.Empty(t, tg.photos)
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "http://link/VID/")
}

func TestSendMediaPhotoErrorIsLogged(t *testing.T) {
	tg := &fakeTelegramClient{err: errors.New("boom")}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}

	// Should not panic; the error is logged and swallowed.
	svc.sendMedia("foo", "IMG", "http://link/IMG/", 0, igMedia{IsVideo: false, URL: "http://img/x"})

	require.Len(t, tg.photos, 1)
	assert.Equal(t, "http://img/x", tg.photos[0].url)
}

const igSkippableFeedJSON = `{"items":[
	{"code":"NOVID","media_type":2},
	{"code":"NOIMG","media_type":1},
	{"code":"EMPTYCAR","media_type":8,"carousel_media":[{"media_type":2}]}
]}`

func TestFetchLatestPostsSkipsMediaWithoutURLs(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte(igSkippableFeedJSON), nil
	}}
	svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: &fakeInstagramRepo{}}

	posts, err := svc.fetchLatestPosts(repository.InstagramAccount{Username: "foo", UserID: "123"})
	require.NoError(t, err)
	// every item resolves to zero usable media, so nothing is emitted
	assert.Empty(t, posts)
}

