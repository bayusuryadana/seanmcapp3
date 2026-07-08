class InstagramClient
  APP_ID = "936619743392459".freeze
  USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36".freeze

  def get(url)
    uri = URI(url)
    request = Net::HTTP::Get.new(uri)
    request["Cookie"] = "sessionid=#{AppSettings.fetch!("IG_SESSION_ID")}; csrftoken=#{AppSettings.fetch!("IG_CSRF_TOKEN")}"
    request["X-CSRFToken"] = AppSettings.fetch!("IG_CSRF_TOKEN")
    request["X-IG-App-ID"] = APP_ID
    request["Referer"] = "https://www.instagram.com/"
    request["User-Agent"] = USER_AGENT

    response = Net::HTTP.start(uri.host, uri.port, use_ssl: true, read_timeout: 15, open_timeout: 15) do |http|
      http.request(request)
    end

    if [401, 403].include?(response.code.to_i)
      raise "session expired or blocked (HTTP #{response.code}) - please update IG_SESSION_ID in .env"
    end
    raise "unexpected status #{response.code} for #{url}" unless response.is_a?(Net::HTTPSuccess)

    response.body
  end
end
