class InstagramFetchService
  PROFILE_BASE = "https://www.instagram.com/api/v1/users/web_profile_info/?username=".freeze
  FEED_BASE = "https://www.instagram.com/api/v1/feed/user/".freeze
  POST_BASE = "https://www.instagram.com/p/".freeze
  MAX_POSTS = 9

  Post = Struct.new(:shortcode, :display_url)

  def initialize(instagram_client: InstagramClient.new, telegram_client: TelegramClient.new, personal_chat_id: AppSettings.fetch!("TELEGRAM_PERSONAL_CHAT_ID").to_i)
    @instagram_client = instagram_client
    @telegram_client = telegram_client
    @personal_chat_id = personal_chat_id
  end

  def run
    InstagramAccount.find_each.with_index do |account, index|
      sleep 5 if index.positive?

      Rails.logger.info("Checking Instagram account: #{account.username}")
      posts = fetch_latest_posts(account.username)
      new_posts = detect_new_posts(account.last_shortcodes.to_s, posts)

      if new_posts.any?
        notify(account.username, new_posts)
      else
        Rails.logger.info("No new posts for #{account.username}")
      end

      account.update!(last_shortcodes: posts.map(&:shortcode).join(","))
    rescue StandardError => error
      Rails.logger.error("[ERROR] fetching posts for #{account.username}: #{error.message}")
    end
  end

  private

  def fetch_latest_posts(username)
    profile = JSON.parse(@instagram_client.get(PROFILE_BASE + username))
    user_id = profile.dig("data", "user", "id")
    raise "could not resolve user id for #{username}" if user_id.blank?

    feed = JSON.parse(@instagram_client.get("#{FEED_BASE}#{user_id}/?count=#{MAX_POSTS}"))
    items = feed["items"]
    raise "unexpected feed structure for #{username}" unless items

    items.map do |item|
      shortcode = item["code"].to_s
      next if shortcode.empty?

      Post.new(shortcode, item.dig("image_versions2", "candidates", 0, "url").to_s)
    end.compact
  end

  def detect_new_posts(stored_raw, current_posts)
    return [] if stored_raw.empty?

    stored = stored_raw.split(",").map(&:strip).to_set
    current_posts.reject { |post| stored.include?(post.shortcode) }
  end

  def notify(username, new_posts)
    new_posts.each do |post|
      caption = "📸 New post from *#{username}*\n🔗 #{POST_BASE}#{post.shortcode}/"
      @telegram_client.send_photo(@personal_chat_id, post.display_url, caption)
      sleep 1
    rescue StandardError => error
      Rails.logger.error("[ERROR] sending photo for #{username}/#{post.shortcode}: #{error.message}")
    end
  end
end
