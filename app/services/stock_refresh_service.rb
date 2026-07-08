class StockRefreshService
  def initialize(price_client: StockPriceClient.new)
    @price_client = price_client
  end

  def refresh_prices
    Stock.find_each do |stock|
      current_price = @price_client.get_price(stock.name)
      stock.update!(current_price: current_price)
    rescue StandardError => error
      Rails.logger.error("[ERROR] #{error.message}")
    end
  end
end
