class TelegramClient
  def initialize(endpoint: AppSettings.fetch!("TELEGRAM_BOT_ENDPOINT"))
    @endpoint = endpoint
  end

  def send_message(chat_id, text)
    get_json("sendmessage", chat_id: chat_id, text: text, parse_mode: "markdown", disable_web_page_preview: true, disable_notification: true)
  end

  def send_photo(chat_id, photo_url, caption)
    get_json("sendphoto", chat_id: chat_id, photo: photo_url, caption: caption, parse_mode: "markdown", disable_notification: true)
  end

  private

  def get_json(method, params)
    uri = URI("#{@endpoint}/#{method}")
    uri.query = URI.encode_www_form(params)
    response = Net::HTTP.get_response(uri)
    JSON.parse(response.body)
  rescue JSON::ParserError => error
    Rails.logger.error("Failed to decode telegram #{method} response: #{error.message}")
    raise
  end
end
