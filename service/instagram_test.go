package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"seanmcapp/repository"
)

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
