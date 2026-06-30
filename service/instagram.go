package service

import (
	"fmt"
	"log"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const (
	igProfileBase = "https://www.instagram.com/api/v1/users/web_profile_info/?username="
	igFeedBase    = "https://www.instagram.com/api/v1/feed/user/"
	igPostBase    = "https://www.instagram.com/p/"
	igMaxPosts    = 9
)

type InstagramService interface {
	Run()
}

type InstagramServiceImpl struct {
	InstagramAccountRepo repository.InstagramAccountRepo
	InstagramClient      external.InstagramClient
	TelegramClient       external.TelegramClient
	PersonalChatID       int64
}

type igPost struct {
	Shortcode  string
	DisplayURL string
}

func (s *InstagramServiceImpl) Run() {
	accounts, err := s.InstagramAccountRepo.GetAll()
	if err != nil {
		log.Println("[ERROR] fetching instagram accounts:", err)
		return
	}

	for i, account := range accounts {
		if i > 0 {
			time.Sleep(5 * time.Second)
		}

		log.Printf("Checking Instagram account: %s", account.Username)
		posts, err := s.fetchLatestPosts(account.Username)
		if err != nil {
			log.Printf("[ERROR] fetching posts for %s: %v", account.Username, err)
			continue
		}

		newPosts := detectNewPosts(account.LastShortcodes, posts)

		if len(newPosts) > 0 {
			s.notify(account.Username, newPosts)
		} else {
			log.Printf("No new posts for %s", account.Username)
		}

		shortcodes := make([]string, len(posts))
		for i, p := range posts {
			shortcodes[i] = p.Shortcode
		}
		if err := s.InstagramAccountRepo.UpdateLastShortcodes(account.Username, strings.Join(shortcodes, ",")); err != nil {
			log.Printf("[ERROR] updating shortcodes for %s: %v", account.Username, err)
		}
	}
}

func (s *InstagramServiceImpl) fetchLatestPosts(username string) ([]igPost, error) {
	// Step 1: resolve the numeric user id from the username.
	profileBody, err := s.InstagramClient.Get(igProfileBase + username)
	if err != nil {
		return nil, err
	}

	userID := gjson.GetBytes(profileBody, "data.user.id").String()
	if userID == "" {
		return nil, fmt.Errorf("could not resolve user id for %s", username)
	}

	// Step 2: fetch the actual posts from the user's feed endpoint.
	feedURL := fmt.Sprintf("%s%s/?count=%d", igFeedBase, userID, igMaxPosts)
	feedBody, err := s.InstagramClient.Get(feedURL)
	if err != nil {
		return nil, err
	}

	items := gjson.GetBytes(feedBody, "items")
	if !items.Exists() {
		return nil, fmt.Errorf("unexpected feed structure for %s", username)
	}

	var posts []igPost
	items.ForEach(func(_, item gjson.Result) bool {
		shortcode := item.Get("code").String()
		imageURL := item.Get("image_versions2.candidates.0.url").String()
		if shortcode != "" {
			posts = append(posts, igPost{Shortcode: shortcode, DisplayURL: imageURL})
		}
		return true
	})

	return posts, nil
}

func detectNewPosts(storedRaw string, current []igPost) []igPost {
	if storedRaw == "" {
		return nil
	}

	stored := make(map[string]bool)
	for _, sc := range strings.Split(storedRaw, ",") {
		stored[strings.TrimSpace(sc)] = true
	}

	var newPosts []igPost
	for _, p := range current {
		if !stored[p.Shortcode] {
			newPosts = append(newPosts, p)
		}
	}
	return newPosts
}

func (s *InstagramServiceImpl) notify(username string, newPosts []igPost) {
	for _, p := range newPosts {
		caption := fmt.Sprintf("📸 New post from *%s*\n🔗 %s%s/", username, igPostBase, p.Shortcode)
		if _, err := s.TelegramClient.SendPhoto(s.PersonalChatID, p.DisplayURL, caption); err != nil {
			log.Printf("[ERROR] sending photo for %s/%s: %v", username, p.Shortcode, err)
		}
		time.Sleep(1 * time.Second)
	}
}
