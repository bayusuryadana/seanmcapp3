package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return doc
}

func TestNewsParsers(t *testing.T) {
	tests := []struct {
		name      string
		source    NewsObject
		html      string
		wantTitle string
		wantURL   string
	}{
		{
			name:      "Detik",
			source:    Detik{},
			html:      `<a dtr-evt="headline" dtr-ttl="Title D" href="https://detik.com/1">x</a>`,
			wantTitle: "Title D",
			wantURL:   "https://detik.com/1",
		},
		{
			name:      "Kumparan",
			source:    Kumparan{},
			html:      `<div data-qa-id="news-item"><a href="/news/1">l</a><div data-qa-id="title">Title K</div></div>`,
			wantTitle: "Title K",
			wantURL:   "https://kumparan.com/news/1",
		},
		{
			name:      "CNA",
			source:    CNA{},
			html:      `<div class="card-object"><h3><a href="/sg/1">Title C</a></h3></div>`,
			wantTitle: "Title C",
			wantURL:   "https://www.channelnewsasia.com/sg/1",
		},
		{
			name:   "Tirtol",
			source: Tirtol{},
			html: `<div>
				<div class="mb-3"><a href="/pop/1">Title T</a></div>
				<div><div><div class="welcome-title">POPULER</div></div></div>
			</div>`,
			wantTitle: "Title T",
			wantURL:   "https://tirto.id/pop/1",
		},
		{
			name:      "Mothership",
			source:    Mothership{},
			html:      `<div class="main-item"><div class="top-story"><h1>Title M</h1><a href="https://mothership.sg/1">l</a></div></div>`,
			wantTitle: "Title M",
			wantURL:   "https://mothership.sg/1",
		},
		{
			name:   "Reuters",
			source: Reuters{},
			html: `<div id="main-content"><div>
				<div><a href="/world/">World</a></div>
				<a data-testid="Heading" href="/world/story1">Title R</a>
			</div></div>`,
			wantTitle: "Title R",
			wantURL:   "https://www.reuters.com/world/story1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			title, url, err := tc.source.Parse(mustDoc(t, tc.html))
			require.NoError(t, err)
			assert.Equal(t, tc.wantTitle, strings.TrimSpace(title))
			assert.Equal(t, tc.wantURL, url)

			// Metadata sanity.
			assert.NotEmpty(t, tc.source.Name())
			assert.NotEmpty(t, tc.source.URL())
			assert.NotEmpty(t, tc.source.Flag())
		})
	}
}

func TestNewsParsersErrorOnEmptyDoc(t *testing.T) {
	sources := []NewsObject{Detik{}, Kumparan{}, CNA{}, Tirtol{}, Mothership{}, Reuters{}}
	for _, src := range sources {
		t.Run(src.Name(), func(t *testing.T) {
			_, _, err := src.Parse(mustDoc(t, "<html></html>"))
			assert.Error(t, err)
		})
	}
}

// fakeNewsSource lets us exercise Run/fetchNews against a local test server.
type fakeNewsSource struct {
	url     string
	parseFn func(*goquery.Document) (string, string, error)
}

func (f fakeNewsSource) Name() string { return "Fake" }
func (f fakeNewsSource) URL() string  { return f.url }
func (f fakeNewsSource) Flag() []int  { return []int{0x1f600} }
func (f fakeNewsSource) Parse(doc *goquery.Document) (string, string, error) {
	return f.parseFn(doc)
}

func TestNewsRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><h1>hi</h1></body></html>`))
	}))
	defer srv.Close()

	tg := &fakeTelegramClient{}
	svc := &NewsServiceImpl{
		TelegramClient: tg,
		GroupChatID:    777,
		httpClient:     srv.Client(),
		sources: []NewsObject{
			fakeNewsSource{url: srv.URL, parseFn: func(*goquery.Document) (string, string, error) {
				return "Breaking News", "https://example.com/story", nil
			}},
		},
	}

	svc.Run()

	require.Len(t, tg.messages, 1)
	assert.Equal(t, int64(777), tg.messages[0].chatID)
	assert.Contains(t, tg.messages[0].text, "Breaking News")
	assert.Contains(t, tg.messages[0].text, "https://example.com/story")
}
