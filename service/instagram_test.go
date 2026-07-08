package service

import (
	"errors"
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
	svc := &InstagramServiceImpl{InstagramClient: client}

	posts, err := svc.fetchLatestPosts("foo")
	require.NoError(t, err)
	require.Len(t, posts, 2)
	assert.Equal(t, "AAA", posts[0].Shortcode)
	assert.Equal(t, "http://img/a", posts[0].DisplayURL)
}

func TestFetchLatestPostsErrors(t *testing.T) {
	t.Run("profile request error", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return nil, errors.New("net") },
		}}
		_, err := svc.fetchLatestPosts("foo")
		assert.Error(t, err)
	})

	t.Run("empty user id", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(string) ([]byte, error) { return []byte(`{"data":{"user":{}}}`), nil },
		}}
		_, err := svc.fetchLatestPosts("foo")
		assert.Error(t, err)
	})

	t.Run("missing items in feed", func(t *testing.T) {
		svc := &InstagramServiceImpl{InstagramClient: &fakeInstagramClient{
			getFn: func(url string) ([]byte, error) {
				if strings.Contains(url, "web_profile_info") {
					return []byte(igProfileJSON), nil
				}
				return []byte(`{}`), nil
			},
		}}
		_, err := svc.fetchLatestPosts("foo")
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
