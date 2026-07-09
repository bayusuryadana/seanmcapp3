package service

import (
	"errors"
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

	// igMaxUploadBytes is Telegram's bot multipart upload ceiling. Videos larger
	// than this fall back to a thumbnail with a link.
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

func (s *InstagramServiceImpl) Run() {
	s.guard.run("instagram run", func() {
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
			posts, err := s.fetchLatestPosts(account)
			if err != nil {
				if errors.Is(err, external.ErrSessionExpired) {
					// Every account will fail the same way, so alert once and stop the run.
					log.Printf("[ERROR] instagram session expired while checking %s: %v", account.Username, err)
					if _, sendErr := s.TelegramClient.SendMessage(s.PersonalChatID, "⚠️ Instagram session expired — please update *IG_SESSION_ID*."); sendErr != nil {
						log.Printf("[ERROR] sending session-expired alert: %v", sendErr)
					}
					return
				}
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
	})
}

func (s *InstagramServiceImpl) fetchLatestPosts(account repository.InstagramAccount) ([]igPost, error) {
	// Step 1: resolve the numeric user id from the username, but only when it is
	// not already stored. Once resolved, persist it so future runs can skip this call.
	userID := account.UserID
	if userID == "" {
		profileBody, err := s.InstagramClient.Get(igProfileBase + account.Username)
		if err != nil {
			return nil, err
		}

		userID = gjson.GetBytes(profileBody, "data.user.id").String()
		if userID == "" {
			return nil, fmt.Errorf("could not resolve user id for %s", account.Username)
		}

		if err := s.InstagramAccountRepo.UpdateUserID(account.Username, userID); err != nil {
			log.Printf("[ERROR] updating user_id for %s: %v", account.Username, err)
		}
	}

	// Step 2: fetch the actual posts from the user's feed endpoint.
	feedURL := fmt.Sprintf("%s%s/?count=%d", igFeedBase, userID, igMaxPosts)
	feedBody, err := s.InstagramClient.Get(feedURL)
	if err != nil {
		return nil, err
	}

	items := gjson.GetBytes(feedBody, "items")
	if !items.Exists() {
		return nil, fmt.Errorf("unexpected feed structure for %s", account.Username)
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

// extractMedia flattens a feed item into its individual media pieces, handling
// single images (media_type 1), videos/reels (media_type 2) and carousels /
// albums (media_type 8, whose children each carry their own media_type).
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
			time.Sleep(1 * time.Second)
		}

		summary := fmt.Sprintf("📸 New post from *%s*\n🔗 %s", username, postLink)
		if caption := strings.TrimSpace(p.Caption); caption != "" {
			summary += "\n\n" + escapeMarkdown(caption)
		}
		if _, err := s.TelegramClient.SendMessage(s.PersonalChatID, summary); err != nil {
			log.Printf("[ERROR] sending summary for %s/%s: %v", username, p.Shortcode, err)
		}
		time.Sleep(1 * time.Second)
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

// sendVideo tries the cheap URL-based send first, then falls back to downloading
// and multipart-uploading (up to 50MB), and finally to a thumbnail + link note.
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

// escapeMarkdown neutralises the legacy Telegram markdown control characters so
// an arbitrary Instagram caption cannot break the summary message formatting.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"`", "\\`",
		"[", "\\[",
	)
	return replacer.Replace(s)
}
