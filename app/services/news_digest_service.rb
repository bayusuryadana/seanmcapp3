class NewsDigestService
  NewsSource = Struct.new(:name, :url, :flags, :parser, keyword_init: true)

  SOURCES = [
    NewsSource.new(name: "Detik", url: "https://www.detik.com/", flags: [0x1f1ee, 0x1f1e9], parser: :parse_detik),
    NewsSource.new(name: "Tirtol", url: "https://tirto.id", flags: [0x1f1ee, 0x1f1e9], parser: :parse_tirtol),
    NewsSource.new(name: "Kumparan", url: "https://kumparan.com/trending", flags: [0x1f1ee, 0x1f1e9], parser: :parse_kumparan),
    NewsSource.new(name: "CNA", url: "https://www.channelnewsasia.com/news/singapore", flags: [0x1f1f8, 0x1f1ec], parser: :parse_cna),
    NewsSource.new(name: "Mothership", url: "https://mothership.sg", flags: [0x1f1f8, 0x1f1ec], parser: :parse_mothership),
    NewsSource.new(name: "Reuters", url: "https://www.reuters.com", flags: [0x1f30f], parser: :parse_reuters)
  ].freeze

  def initialize(telegram_client: TelegramClient.new, group_chat_id: AppSettings.fetch!("TELEGRAM_GROUP_CHAT_ID").to_i)
    @telegram_client = telegram_client
    @group_chat_id = group_chat_id
  end

  def run
    results = SOURCES.map do |source|
      fetch_news(source)
    rescue StandardError => error
      Rails.logger.error("[ERROR] #{source.name}: #{error.message}")
      nil
    end.compact

    message = "Awali harimu dengan berita 📰 dari **Seanmctoday** by @seanmcbot\n\n"
    results.each do |result|
      message += "#{result[:flags]} #{result[:source]} - [#{result[:title].strip}](#{result[:url]})\n\n"
    end

    @telegram_client.send_message(@group_chat_id, message)
  end

  private

  def fetch_news(source)
    response = Net::HTTP.get_response(URI(source.url))
    document = Nokogiri::HTML(response.body)
    title, url = send(source.parser, document)

    { source: source.name, flags: source.flags.pack("U*"), title: title, url: url }
  end

  def parse_detik(document)
    tag = document.css("[dtr-evt=headline]").first
    raise "headline not found for Detik" unless tag

    [tag["dtr-ttl"].to_s, tag["href"].to_s]
  end

  def parse_kumparan(document)
    tag = document.css("[data-qa-id=news-item]").first
    raise "news item not found for Kumparan" unless tag

    title = tag.css("[data-qa-id=title]").text
    href = tag.css("a").first&.[]("href").to_s
    raise "missing title or link for Kumparan" if title.empty? || href.empty?

    [title, "https://kumparan.com#{href}"]
  end

  def parse_cna(document)
    tag = document.css(".card-object h3").first
    raise "CNA card title not found" unless tag

    title = tag.text
    href = tag.css("a").first&.[]("href").to_s
    raise "missing title or link for CNA" if title.empty? || href.empty?

    [title, "https://www.channelnewsasia.com#{href}"]
  end

  def parse_tirtol(document)
    title = document.css(".welcome-title").find { |node| node.text.strip == "POPULER" }
    tag = title&.parent&.parent&.parent&.css(".mb-3 a")&.first
    raise "POPULAR section not found for Tirtol" unless tag

    [tag.text, "https://tirto.id#{tag["href"]}"]
  end

  def parse_mothership(document)
    tag = document.css(".main-item > .top-story").first
    raise "top story not found for Mothership" unless tag

    title = tag.css("h1").text
    href = tag.css("a").first&.[]("href").to_s
    raise "missing title or link for Mothership" if title.empty? || href.empty?

    [title, href]
  end

  def parse_reuters(document)
    tag = document.css("#main-content [href='/world/']").first&.parent&.parent&.css("[data-testid=Heading]")&.first
    raise "Reuters headline not found" unless tag

    title = tag.text
    href = tag["href"].to_s
    raise "missing title or link for Reuters" if title.empty? || href.empty?

    [title, "https://www.reuters.com#{href}"]
  end
end
