class AppSettings
  REQUIRED_KEYS = %w[
    DATABASE_HOST
    DATABASE_NAME
    DATABASE_PASS
    DATABASE_USER
    APPS_SECRET_KEY
    APPS_PASSWORD
    TELEGRAM_BOT_ENDPOINT
    TELEGRAM_BOT_NAME
    TELEGRAM_PERSONAL_CHAT_ID
    TELEGRAM_GROUP_CHAT_ID
    IG_SESSION_ID
    IG_CSRF_TOKEN
  ].freeze

  def self.fetch!(key)
    ENV.fetch(key) { raise KeyError, "#{key} is not set" }
  end

  def self.validate!
    REQUIRED_KEYS.each { |key| fetch!(key) }
  end
end
