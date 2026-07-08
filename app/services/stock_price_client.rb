class StockPriceClient
  URL_TEMPLATE = "https://query1.finance.yahoo.com/v8/finance/chart/%{name}.jk".freeze
  USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36".freeze

  def get_price(name)
    uri = URI(format(URL_TEMPLATE, name: name))
    request = Net::HTTP::Get.new(uri)
    request["User-Agent"] = USER_AGENT

    response = Net::HTTP.start(uri.host, uri.port, use_ssl: true, read_timeout: 15, open_timeout: 15) do |http|
      http.request(request)
    end

    price = JSON.parse(response.body).dig("chart", "result", 0, "meta", "regularMarketPrice")
    raise "stock #{name} not found in json" if price.nil?

    price.to_i
  end
end
