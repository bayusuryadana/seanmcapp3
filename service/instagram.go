package service

import (
	"fmt"
	"log"
	"math/rand"
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

	igMaxUploadBytes = 50 * 1024 * 1024
)

type InstagramService interface {
	Run()
}

type InstagramServiceImpl struct {
	InstagramAccountRepo repository.InstagramAccountRepo
	InstagramClient      external.InstagramClient
	TelegramClient       external.TelegramClient
	PersonalChatID       int64
	guard                runGuard
}

type igMedia struct {
	IsVideo      bool
	URL          string
	ThumbnailURL string
}

type igPost struct {
	Shortcode string
	Caption   string
	Media     []igMedia
}

var sleepFn = time.Sleep
var hourFn = func() int { return time.Now().Hour() }

func init() {
	rand.Seed(time.Now().UnixNano())
}

func selectAccountsForHour(accounts []repository.InstagramAccount, hour int) []repository.InstagramAccount {
	if len(accounts) == 0 {
		return nil
	}

	bucketCount := 24
	if len(accounts) < bucketCount {
		bucketCount = len(accounts)
	}

	selected := make([]repository.InstagramAccount, 0)
	target := (hour + 1) % bucketCount
	for _, account := range accounts {
		if account.ID%bucketCount == target {
			selected = append(selected, account)
		}
	}
	return selected
}

func randomDuration(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

func sleepRandom(min, max time.Duration) {
	sleepFn(randomDuration(min, max))
}

func (s *InstagramServiceImpl) Run() {
	s.guard.run("instagram run", func() {
		accounts, err := s.InstagramAccountRepo.GetAll()
		if err != nil {
			log.Println("[ERROR] fetching instagram accounts:", err)
			return
		}

		accounts = selectAccountsForHour(accounts, hourFn())
		if len(accounts) == 0 {
			log.Println("[INFO] no Instagram accounts selected for this hour")
			return
		}

		sleepRandom(30*time.Second, 90*time.Second)

		for i, account := range accounts {
			if i > 0 {
				sleepRandom(5*time.Second, 12*time.Second)
			}

			log.Printf("Checking Instagram account: %s", account.Username)

			s.processAccount(account)
		}

		log.Printf("===== Instagram run/trigger is completed =====")
	})
}

func (s *InstagramServiceImpl) processAccount(account repository.InstagramAccount) {
	userID, err := s.resolveUserID(account)
	if err != nil {
		log.Printf("[ERROR] resolving user id for %s: %v", account.Username, err)
		return
	}

	s.processPosts(account, userID)
}

func (s *InstagramServiceImpl) processPosts(account repository.InstagramAccount, userID string) {
	posts, err := s.fetchLatestPosts(account.Username, userID)
	if err != nil {
		log.Printf("[ERROR] fetching posts for %s: %v", account.Username, err)
		return
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

func (s *InstagramServiceImpl) resolveUserID(account repository.InstagramAccount) (string, error) {
	if account.UserID != "" {
		return account.UserID, nil
	}

	profileBody, err := s.InstagramClient.Get(igProfileBase + account.Username)
	if err != nil {
		return "", err
	}

	userID := gjson.GetBytes(profileBody, "data.user.id").String()
	if userID == "" {
		return "", fmt.Errorf("could not resolve user id for %s", account.Username)
	}

	if err := s.InstagramAccountRepo.UpdateUserID(account.Username, userID); err != nil {
		log.Printf("[ERROR] updating user_id for %s: %v", account.Username, err)
	}
	return userID, nil
}

func (s *InstagramServiceImpl) fetchLatestPosts(username, userID string) ([]igPost, error) {
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
		if shortcode == "" {
			return true
		}
		media := extractMedia(item)
		if len(media) == 0 {
			return true
		}
		posts = append(posts, igPost{
			Shortcode: shortcode,
			Caption:   item.Get("caption.text").String(),
			Media:     media,
		})
		return true
	})

	return posts, nil
}

func extractMedia(item gjson.Result) []igMedia {
	if item.Get("media_type").Int() == 8 {
		var media []igMedia
		item.Get("carousel_media").ForEach(func(_, child gjson.Result) bool {
			if m, ok := singleMedia(child); ok {
				media = append(media, m)
			}
			return true
		})
		return media
	}

	if m, ok := singleMedia(item); ok {
		return []igMedia{m}
	}
	return nil
}

func singleMedia(node gjson.Result) (igMedia, bool) {
	thumb := node.Get("image_versions2.candidates.0.url").String()

	if node.Get("media_type").Int() == 2 {
		videoURL := node.Get("video_versions.0.url").String()
		if videoURL == "" {
			return igMedia{}, false
		}
		return igMedia{IsVideo: true, URL: videoURL, ThumbnailURL: thumb}, true
	}

	if thumb == "" {
		return igMedia{}, false
	}
	return igMedia{IsVideo: false, URL: thumb}, true
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
		postLink := fmt.Sprintf("%s%s/", igPostBase, p.Shortcode)

		for i, m := range p.Media {
			s.sendMedia(username, p.Shortcode, postLink, i, m)
			sleepRandom(1100*time.Millisecond, 2600*time.Millisecond)
		}

		summary := fmt.Sprintf("📸 New post from *%s*\n🔗 [%s](%s)", escapeMarkdown(username), escapeMarkdown(postLink), postLink)
		if caption := strings.TrimSpace(p.Caption); caption != "" {
			summary += "\n\n" + escapeMarkdown(caption)
		}
		if _, err := s.TelegramClient.SendMessage(s.PersonalChatID, summary); err != nil {
			log.Printf("[ERROR] sending summary for %s/%s: %v", username, p.Shortcode, err)
		}
		sleepRandom(1100*time.Millisecond, 2700*time.Millisecond)
	}
}

func (s *InstagramServiceImpl) sendMedia(username, shortcode, postLink string, index int, m igMedia) {
	if !m.IsVideo {
		if _, err := s.TelegramClient.SendPhoto(s.PersonalChatID, m.URL, ""); err != nil {
			log.Printf("[ERROR] sending photo for %s/%s: %v", username, shortcode, err)
		}
		return
	}
	s.sendVideo(username, shortcode, postLink, index, m)
}

func (s *InstagramServiceImpl) sendVideo(username, shortcode, postLink string, index int, m igMedia) {
	if resp, err := s.TelegramClient.SendVideo(s.PersonalChatID, m.URL, ""); err == nil && resp.Ok {
		return
	}

	data, err := s.InstagramClient.Get(m.URL)
	if err != nil {
		log.Printf("[ERROR] downloading video for %s/%s: %v", username, shortcode, err)
		s.sendVideoFallback(username, shortcode, postLink, m)
		return
	}
	if len(data) > igMaxUploadBytes {
		s.sendVideoFallback(username, shortcode, postLink, m)
		return
	}

	filename := fmt.Sprintf("%s_%d.mp4", shortcode, index)
	if resp, err := s.TelegramClient.SendVideoUpload(s.PersonalChatID, data, filename, ""); err != nil || !resp.Ok {
		log.Printf("[ERROR] uploading video for %s/%s (ok=%t): %v", username, shortcode, resp.Ok, err)
		s.sendVideoFallback(username, shortcode, postLink, m)
	}
}

func (s *InstagramServiceImpl) sendVideoFallback(username, shortcode, postLink string, m igMedia) {
	note := fmt.Sprintf("🎬 This one's a video — too big to preview here. Watch it on Instagram 👉 %s", postLink)
	if m.ThumbnailURL == "" {
		if _, err := s.TelegramClient.SendMessage(s.PersonalChatID, note); err != nil {
			log.Printf("[ERROR] sending video fallback note for %s/%s: %v", username, shortcode, err)
		}
		return
	}
	if _, err := s.TelegramClient.SendPhoto(s.PersonalChatID, m.ThumbnailURL, note); err != nil {
		log.Printf("[ERROR] sending video fallback for %s/%s: %v", username, shortcode, err)
	}
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"`", "\\`",
		"[", "\\[",
	)
	return replacer.Replace(s)
}
