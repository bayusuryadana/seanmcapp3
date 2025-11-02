package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"seanmcapp/external"
	"seanmcapp/util"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type NewsService interface {
	Run()
}

type NewsServiceImpl struct {
	TelegramClient external.TelegramClient
}

func (s *NewsServiceImpl) Run() {
	newsList := []NewsObject{
		Detik{},
		Tirtol{},
		Kumparan{},
		CNA{},
		Mothership{},
		Reuters{},
	}

	var results []NewsResult

	for _, news := range newsList {
		resp, err := http.Get(news.URL())
		if err != nil {
			log.Printf("[ERROR] fetching %s: %v\n", news.Name(), err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[ERROR] reading response: %v\n", err)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			log.Printf("[ERROR] parsing document: %v\n", err)
			continue
		}

		title, url, err := news.Parse(doc)
		if err != nil {
			log.Printf("[ERROR] parsing news from %s: %v\n", news.Name(), err)
			continue
		}

		results = append(results, NewsResult{
			Title:      title,
			URL:        url,
			NewsSource: news,
		})
	}

	message := "Awali harimu dengan berita 📰 dari **Seanmctoday** by @seanmcbot\n\n"
	for _, res := range results {
		flags := ""
		for _, f := range res.NewsSource.Flag() {
			flags += string(rune(f))
		}
		message += fmt.Sprintf("%s %s - [%s](%s)\n\n", flags, res.NewsSource.Name(), strings.TrimSpace(res.Title), res.URL)
	}

	groupChatId := util.GetAppSettings().TelegramSettings.GroupChatID
	s.TelegramClient.SendMessage(groupChatId, message)
}

type NewsObject interface {
	Name() string
	URL() string
	Flag() []int
	Parse(doc *goquery.Document) (string, string, error)
}

type NewsResult struct {
	Title      string
	URL        string
	NewsSource NewsObject
}

type Detik struct{}

func (d Detik) Name() string { return "Detik" }
func (d Detik) URL() string  { return "https://www.detik.com/" }
func (d Detik) Flag() []int  { return []int{0x1f1ee, 0x1f1e9} }

func (d Detik) Parse(doc *goquery.Document) (string, string, error) {
	tag := doc.Find("[dtr-evt=headline]").First()
	if tag.Length() == 0 {
		return "", "", errors.New("headline not found for Detik")
	}
	return tag.AttrOr("dtr-ttl", ""), tag.AttrOr("href", ""), nil
}

type Kumparan struct{}

func (k Kumparan) Name() string { return "Kumparan" }
func (k Kumparan) URL() string  { return "https://kumparan.com/trending" }
func (k Kumparan) Flag() []int  { return []int{0x1f1ee, 0x1f1e9} }
func (k Kumparan) Parse(doc *goquery.Document) (string, string, error) {
	tag := doc.Find("[data-qa-id=news-item]").First()
	if tag.Length() == 0 {
		return "", "", fmt.Errorf("news item not found for Kumparan")
	}

	title := tag.Find("[data-qa-id=title]").Text()
	href := tag.Find("a").AttrOr("href", "")

	if title == "" || href == "" {
		return "", "", fmt.Errorf("missing title or link for Kumparan")
	}
	return title, "https://kumparan.com" + href, nil
}

type CNA struct{}

func (c CNA) Name() string { return "CNA" }
func (c CNA) URL() string  { return "https://www.channelnewsasia.com/news/singapore" }
func (c CNA) Flag() []int  { return []int{0x1f1f8, 0x1f1ec} }
func (c CNA) Parse(doc *goquery.Document) (string, string, error) {
	tag := doc.Find(".card-object h3").First()
	if tag.Length() == 0 {
		return "", "", fmt.Errorf("CNA card title not found")
	}

	title := tag.Text()
	href := tag.Find("a").AttrOr("href", "")
	if title == "" || href == "" {
		return "", "", fmt.Errorf("missing title or link for CNA")
	}
	return title, "https://www.channelnewsasia.com" + href, nil
}

type Tirtol struct{}

func (t Tirtol) Name() string { return "Tirtol" }
func (t Tirtol) URL() string  { return "https://tirto.id" }
func (t Tirtol) Flag() []int  { return []int{0x1f1ee, 0x1f1e9} }
func (t Tirtol) Parse(doc *goquery.Document) (string, string, error) {
	var found bool
	var tag *goquery.Selection

	doc.Find(".welcome-title").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.TrimSpace(s.Text()) == "POPULER" {
			tag = s.Parent().Parent().Parent().Find(".mb-3 a").First()
			found = true
			return false
		}
		return true
	})

	if !found || tag.Length() == 0 {
		return "", "", fmt.Errorf("POPULAR section not found for Tirtol")
	}
	return tag.Text(), "https://tirto.id" + tag.AttrOr("href", ""), nil
}

type Mothership struct{}

func (m Mothership) Name() string { return "Mothership" }
func (m Mothership) URL() string  { return "https://mothership.sg" }
func (m Mothership) Flag() []int  { return []int{0x1f1f8, 0x1f1ec} }
func (m Mothership) Parse(doc *goquery.Document) (string, string, error) {
	tag := doc.Find(".main-item > .top-story").First()
	if tag.Length() == 0 {
		return "", "", fmt.Errorf("top story not found for Mothership")
	}

	title := tag.Find("h1").Text()
	href := tag.Find("a").AttrOr("href", "")

	if title == "" || href == "" {
		return "", "", fmt.Errorf("missing title or link for Mothership")
	}
	return title, href, nil
}

type Reuters struct{}

func (r Reuters) Name() string { return "Reuters" }
func (r Reuters) URL() string  { return "https://www.reuters.com" }
func (r Reuters) Flag() []int  { return []int{0x1f30f} }
func (r Reuters) Parse(doc *goquery.Document) (string, string, error) {
	main := doc.Find("#main-content").First()
	tag := main.Find("[href='/world/']").Parent().Parent().Find("[data-testid=Heading]").First()
	if tag.Length() == 0 {
		return "", "", fmt.Errorf("Reuters headline not found")
	}

	title := tag.Text()
	href := tag.AttrOr("href", "")
	if title == "" || href == "" {
		return "", "", fmt.Errorf("missing title or link for Reuters")
	}
	return title, "https://www.reuters.com" + href, nil
}
