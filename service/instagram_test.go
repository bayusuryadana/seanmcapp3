package service

import (
	"errors"
	"fmt"
	"os"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	sleepFn = func(time.Duration) {}
	os.Exit(m.Run())
}

func TestDetectNewPosts(t *testing.T) {
	tests := []struct {
		name    string
		stored  string
		current []igPost
		want    []string
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

func TestResolveUserID(t *testing.T) {
	t.Run("resolves and persists when empty", func(t *testing.T) {
		client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
			return []byte(igProfileJSON), nil
		}}
		repo := &fakeInstagramRepo{}
		svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: repo}

		id, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
		require.NoError(t, err)
		assert.Equal(t, "123", id)
		assert.Equal(t, "123", repo.updatedUserIDs["foo"])
	})

	t.Run("logs update error but returns user id", func(t *testing.T) {
		client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
			return []byte(igProfileJSON), nil
		}}
		repo := &fakeInstagramRepo{updateFn: func(username, shortcodes string) error {
			return errors.New("update failed")
		}}
		svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: repo}

		id, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
		require.NoError(t, err)
		assert.Equal(t, "123", id)
	})

	t.Run("skips profile call when already known", func(t *testing.T) {
		var called []string
		client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
			called = append(called, url)
			return nil, nil
		}}
		repo := &fakeInstagramRepo{}
		svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: repo}

		id, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo", UserID: "999"})
		require.NoError(t, err)
		assert.Equal(t, "999", id)
		assert.Empty(t, called)
		assert.Empty(t, repo.updatedUserIDs)
	})

	t.Run("profile request error", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return nil, errors.New("net") },
		}}
		_, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
		assert.Error(t, err)
	})

	t.Run("empty user id", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return []byte(`{"data":{"user":{}}}`), nil },
		}}
		_, err := svc.resolveUserID(repository.InstagramAccount{Username: "foo"})
		assert.Error(t, err)
	})
}

func TestFetchLatestPosts(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte(igFeedJSON), nil
	}}
	svc := &InstagramServiceImpl{InstagramClient: client, InstagramAccountRepo: &fakeInstagramRepo{}}

	posts, err := svc.fetchLatestPosts("foo", "123")
	require.NoError(t, err)
	require.Len(t, posts, 2)
	assert.Equal(t, "AAA", posts[0].Shortcode)
	require.Len(t, posts[0].Media, 1)
	assert.False(t, posts[0].Media[0].IsVideo)
	assert.Equal(t, "http://img/a", posts[0].Media[0].URL)
}

func TestFetchLatestPostsErrors(t *testing.T) {
	t.Run("feed request error", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return nil, errors.New("net") },
		}}
		_, err := svc.fetchLatestPosts("foo", "123")
		assert.Error(t, err)
	})

	t.Run("missing items in feed", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return []byte(`{}`), nil },
		}}
		_, err := svc.fetchLatestPosts("foo", "123")
		assert.Error(t, err)
	})
}

func TestRandomDurationMinMax(t *testing.T) {
	assert.Equal(t, 100*time.Millisecond, randomDuration(100*time.Millisecond, 100*time.Millisecond))
}

func TestSelectAccountsForHour(t *testing.T) {
	accounts := make([]repository.InstagramAccount, 26)
	for i := range accounts {
		accounts[i].ID = i + 1
		accounts[i].Username = fmt.Sprintf("acct-%d", i)
	}

	t.Run("matches the requested bucket by account ID after hour offset", func(t *testing.T) {
		selected := selectAccountsForHour(accounts, 3)
		require.Len(t, selected, 1)
		assert.Equal(t, "acct-3", selected[0].Username)
	})

	t.Run("maps hour 0 to all matching IDs for target 1", func(t *testing.T) {
		selected := selectAccountsForHour(accounts, 0)
		require.Len(t, selected, 2)
		assert.Equal(t, "acct-0", selected[0].Username)
		assert.Equal(t, "acct-24", selected[1].Username)
	})
}

func TestProcessAccountSwallowsNonSessionErrors(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, igFeedBase) {
			return nil, errors.New("fetch feed failed")
		}
		return []byte(igStoriesJSON), nil
	}}
	repo := &fakeInstagramRepo{}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: repo, InstagramClient: client, TelegramClient: tg}

	err := svc.processAccount(repository.InstagramAccount{Username: "foo", UserID: "123"})
	require.NoError(t, err)
	assert.Equal(t, "111,222", repo.updatedStoryIDs["foo"])
}

func TestProcessAccountExpiresWhenStoriesSessionExpired(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, igStoriesBase) {
			return nil, fmt.Errorf("%w (HTTP 401)", external.ErrSessionExpired)
		}
		return []byte(igFeedJSON), nil
	}}
	repo := &fakeInstagramRepo{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: repo, InstagramClient: client}

	err := svc.processAccount(repository.InstagramAccount{Username: "foo", UserID: "123"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, external.ErrSessionExpired))
}

func TestNotifySendsPhotoSummary(t *testing.T) {
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}
	post := igPost{Shortcode: "IMG", Caption: "hello", Media: []igMedia{{IsVideo: false, URL: "http://img/x"}}}

	svc.notify("foo", []igPost{post})

	require.Len(t, tg.photos, 1)
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "foo")
	assert.Contains(t, tg.messages[0].text, "IMG")
}

func TestSendVideoFallbackThumbnailErrorLogsMessage(t *testing.T) {
	tg := &fakeTelegramClient{err: errors.New("boom")}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}

	svc.sendVideoFallback("foo", "VID", "http://link/VID/", igMedia{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"})

	require.Len(t, tg.photos, 1)
	assert.Equal(t, "http://img/t", tg.photos[0].url)
}

func TestInstagramRun(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{{Username: "foo", LastShortcodes: "AAA,BBB"}}, nil
	}}
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

func TestInstagramRunRefreshesCredentialsEveryTime(t *testing.T) {
	configRepo := &fakeConfigRepo{values: map[string]string{
		repository.ConfigKeyIGSessionID: "session-1",
		repository.ConfigKeyIGCSRFToken: "csrf-1",
	}}
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return nil, nil
	}}
	client := &fakeInstagramClient{}
	svc := &InstagramServiceImpl{
		InstagramAccountRepo: accountRepo,
		ConfigRepo:           configRepo,
		InstagramClient:      client,
	}

	svc.Run()
	assert.Equal(t, "session-1", client.sessionID)
	assert.Equal(t, "csrf-1", client.csrfToken)

	configRepo.values[repository.ConfigKeyIGSessionID] = "session-2"
	configRepo.values[repository.ConfigKeyIGCSRFToken] = "csrf-2"
	svc.Run()

	assert.Equal(t, "session-2", client.sessionID)
	assert.Equal(t, "csrf-2", client.csrfToken)
	assert.Equal(t, 2, client.setCredsRun)
	require.Len(t, configRepo.getCalls, 2)
	assert.Equal(t, []string{
		repository.ConfigKeyIGSessionID,
		repository.ConfigKeyIGCSRFToken,
	}, configRepo.getCalls[0])
}

func TestInstagramRunStopsWhenCredentialsCannotBeLoaded(t *testing.T) {
	accountRepoCalls := 0
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		accountRepoCalls++
		return nil, nil
	}}
	client := &fakeInstagramClient{}

	t.Run("database error", func(t *testing.T) {
		svc := &InstagramServiceImpl{
			InstagramAccountRepo: accountRepo,
			ConfigRepo:           &fakeConfigRepo{getErr: errors.New("database unavailable")},
			InstagramClient:      client,
		}
		svc.Run()
		assert.Equal(t, 0, accountRepoCalls)
		assert.Equal(t, 0, client.setCredsRun)
	})

	t.Run("missing value", func(t *testing.T) {
		svc := &InstagramServiceImpl{
			InstagramAccountRepo: accountRepo,
			ConfigRepo: &fakeConfigRepo{values: map[string]string{
				repository.ConfigKeyIGSessionID: "session",
			}},
			InstagramClient: client,
		}
		svc.Run()
		assert.Equal(t, 0, accountRepoCalls)
		assert.Equal(t, 0, client.setCredsRun)
	})
}

func TestInstagramRunSendsNewPosts(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{{Username: "foo", LastShortcodes: "AAA"}}, nil
	}}
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
	assert.Empty(t, tg.photos[0].caption)
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "foo")
	assert.Contains(t, tg.messages[0].text, "BBB")
	assert.Equal(t, "AAA,BBB", accountRepo.updatedShortcodes["foo"])
}

func TestInstagramRunAlertsOnceOnExpiredSession(t *testing.T) {
	originalHourFn := hourFn
	hourFn = func() int { return 0 }
	t.Cleanup(func() { hourFn = originalHourFn })

	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{
			{ID: 1, Username: "foo", UserID: "1"},
			{ID: 3, Username: "bar", UserID: "2"},
		}, nil
	}}
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return nil, fmt.Errorf("%w (HTTP 401)", external.ErrSessionExpired)
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg, PersonalChatID: 42}

	svc.Run()

	require.Len(t, tg.messages, 1)
	assert.Equal(t, int64(42), tg.messages[0].chatID)
	assert.Contains(t, tg.messages[0].text, "change <session_id> <csrf_token>")
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

	posts, err := svc.fetchLatestPosts("foo", "123")
	require.NoError(t, err)
	require.Len(t, posts, 2)

	require.Len(t, posts[0].Media, 1)
	assert.True(t, posts[0].Media[0].IsVideo)
	assert.Equal(t, "http://vid/v", posts[0].Media[0].URL)
	assert.Equal(t, "http://img/vthumb", posts[0].Media[0].ThumbnailURL)
	assert.Equal(t, "a *fancy* caption", posts[0].Caption)

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
	assert.Empty(t, tg.uploads)
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "foo")
	assert.Contains(t, tg.messages[0].text, "\\_world\\_")
}

func TestNotifyVideoUploadFallbackWhenURLFails(t *testing.T) {
	tg := &fakeTelegramClient{videoURLFails: true}
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		return []byte("small-video-bytes"), nil
	}}
	svc := &InstagramServiceImpl{TelegramClient: tg, InstagramClient: client, PersonalChatID: 7}
	post := igPost{Shortcode: "VID", Media: []igMedia{{IsVideo: true, URL: "http://vid/v", ThumbnailURL: "http://img/t"}}}

	svc.notify("foo", []igPost{post})

	require.Len(t, tg.videos, 1)
	require.Len(t, tg.uploads, 1)
	assert.Equal(t, "VID_0.mp4", tg.uploads[0].filename)
	assert.Empty(t, tg.photos)
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

	assert.Empty(t, tg.uploads)
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

	require.Len(t, tg.uploads, 1)
	require.Len(t, tg.photos, 1)
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

	posts, err := svc.fetchLatestPosts("foo", "123")
	require.NoError(t, err)
	assert.Empty(t, posts)
}

const igStoriesJSON = `{"reels_media":[{"items":[
	{"pk":"111","media_type":1,"image_versions2":{"candidates":[{"url":"http://img/s1"}]}},
	{"pk":"222","media_type":2,"caption":{"text":"story _cap_"},"video_versions":[{"url":"http://vid/s2"}],"image_versions2":{"candidates":[{"url":"http://img/s2thumb"}]}},
	{"pk":"333","media_type":1}
]}]}`

const igStoriesAltShapeJSON = `{"reels":{"123":{"items":[
	{"pk":"777","media_type":1,"image_versions2":{"candidates":[{"url":"http://img/s7"}]}}
]}}}`

func TestFetchLatestStories(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return []byte(igStoriesJSON), nil
	}}
	svc := &InstagramServiceImpl{InstagramClient: client}

	stories, err := svc.fetchLatestStories("foo", "123")
	require.NoError(t, err)
	require.Len(t, stories, 2)
	assert.Equal(t, "111", stories[0].ID)
	assert.False(t, stories[0].Media.IsVideo)
	assert.Equal(t, "http://img/s1", stories[0].Media.URL)
	assert.Equal(t, "222", stories[1].ID)
	assert.True(t, stories[1].Media.IsVideo)
	assert.Equal(t, "http://vid/s2", stories[1].Media.URL)
	assert.Equal(t, "story _cap_", stories[1].Caption)
}

func TestFetchLatestStoriesAltShape(t *testing.T) {
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return []byte(igStoriesAltShapeJSON), nil
	}}
	svc := &InstagramServiceImpl{InstagramClient: client}

	stories, err := svc.fetchLatestStories("foo", "123")
	require.NoError(t, err)
	require.Len(t, stories, 1)
	assert.Equal(t, "777", stories[0].ID)
}

func TestProcessStoriesEmptyCacheSendsEverything(t *testing.T) {
	accountRepo := &fakeInstagramRepo{}
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return []byte(igStoriesJSON), nil
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg}

	account := repository.InstagramAccount{Username: "foo", UserID: "123", LastStoryIDs: ""}
	err := svc.processStories(account, "123")
	require.NoError(t, err)

	require.Len(t, tg.messages, 2)
	assert.Equal(t, "111,222", accountRepo.updatedStoryIDs["foo"])
}

func TestProcessStoriesNoActiveStoryClearsCache(t *testing.T) {
	accountRepo := &fakeInstagramRepo{}
	client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
		return []byte(`{"reels_media":[]}`), nil
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg}

	account := repository.InstagramAccount{Username: "foo", UserID: "123", LastStoryIDs: "111,222"}
	err := svc.processStories(account, "123")
	require.NoError(t, err)

	assert.Equal(t, "", accountRepo.updatedStoryIDs["foo"])
	assert.Empty(t, tg.messages)
}

func TestNotifyStoriesEscapesUnderscoreUsername(t *testing.T) {
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}
	stories := []igStory{{ID: "999", Media: igMedia{URL: "http://img/s"}}}

	svc.notifyStories("jjuya_o0o", stories)

	require.Len(t, tg.messages, 1)
	txt := tg.messages[0].text
	assert.Contains(t, txt, "*jjuya\\_o0o*")
	assert.Contains(t, txt, "[https://www.instagram.com/stories/jjuya\\_o0o/999/]")
	assert.Contains(t, txt, "(https://www.instagram.com/stories/jjuya_o0o/999/)")
}

func TestFetchLatestStoriesEmptyAndErrors(t *testing.T) {
	t.Run("no active reel", func(t *testing.T) {
		client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
			return []byte(`{"reels_media":[]}`), nil
		}}
		svc := &InstagramServiceImpl{InstagramClient: client}
		stories, err := svc.fetchLatestStories("foo", "123")
		require.NoError(t, err)
		assert.Empty(t, stories)
	})

	t.Run("request error", func(t *testing.T) {
		client := &fakeInstagramClient{getFn: func(string) ([]byte, error) {
			return nil, errors.New("net")
		}}
		svc := &InstagramServiceImpl{InstagramClient: client}
		_, err := svc.fetchLatestStories("foo", "123")
		assert.Error(t, err)
	})
}

func TestDetectNewStories(t *testing.T) {
	current := []igStory{{ID: "111"}, {ID: "222"}, {ID: "333"}}

	assert.Len(t, detectNewStories("", current), 3)

	got := detectNewStories("111,222", current)
	require.Len(t, got, 1)
	assert.Equal(t, "333", got[0].ID)

	assert.Empty(t, detectNewStories("111,222,333", current))
}

func TestNotifyStoriesSendsMediaThenSummary(t *testing.T) {
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{TelegramClient: tg, PersonalChatID: 7}
	stories := []igStory{
		{ID: "111", Media: igMedia{IsVideo: false, URL: "http://img/s1"}},
		{ID: "222", Caption: "hi _there_", Media: igMedia{IsVideo: true, URL: "http://vid/s2", ThumbnailURL: "http://img/t"}},
	}

	svc.notifyStories("foo", stories)

	require.Len(t, tg.photos, 1)
	assert.Equal(t, "http://img/s1", tg.photos[0].url)
	require.Len(t, tg.videos, 1)
	assert.Equal(t, "http://vid/s2", tg.videos[0].url)
	require.Len(t, tg.messages, 2)
	assert.Contains(t, tg.messages[0].text, "👀 New story from *foo*")
	assert.Contains(t, tg.messages[0].text, "/stories/foo/111/")
	assert.Contains(t, tg.messages[1].text, "\\_there\\_")
}

func TestInstagramRunSendsNewStory(t *testing.T) {
	accountRepo := &fakeInstagramRepo{getAllFn: func() ([]repository.InstagramAccount, error) {
		return []repository.InstagramAccount{{Username: "foo", UserID: "123", LastShortcodes: "AAA,BBB", LastStoryIDs: "111"}}, nil
	}}
	client := &fakeInstagramClient{getFn: func(url string) ([]byte, error) {
		if strings.Contains(url, "reels_media") {
			return []byte(igStoriesJSON), nil
		}
		return []byte(igFeedJSON), nil
	}}
	tg := &fakeTelegramClient{}
	svc := &InstagramServiceImpl{InstagramAccountRepo: accountRepo, InstagramClient: client, TelegramClient: tg, PersonalChatID: 42}

	svc.Run()

	require.Len(t, tg.videos, 1)
	assert.Equal(t, "http://vid/s2", tg.videos[0].url)
	require.Len(t, tg.messages, 1)
	assert.Contains(t, tg.messages[0].text, "New story from *foo*")
	assert.Equal(t, "111,222", accountRepo.updatedStoryIDs["foo"])
	assert.Empty(t, tg.photos)
}
