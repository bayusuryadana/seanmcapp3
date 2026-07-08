class StockAlertService
  def initialize(telegram_client: TelegramClient.new, personal_chat_id: AppSettings.fetch!("TELEGRAM_PERSONAL_CHAT_ID").to_i)
    @telegram_client = telegram_client
    @personal_chat_id = personal_chat_id
  end

  def run
    StockRefreshService.new.refresh_prices

    result = Stock.all.each_with_object([]) do |stock, messages|
      next if stock.current_price.nil?

      messages << "#{stock.name} hitting best price" if !stock.status && stock.current_price <= stock.best_price
      messages << "#{stock.name} reaching fair price" if stock.status && stock.current_price >= stock.fair_price
    end

    return if result.empty?

    Rails.logger.info("[INFO] stocks hit/reach")
    @telegram_client.send_message(@personal_chat_id, result.join("\n"))
  rescue StandardError => error
    Rails.logger.error("[ERROR] cannot send stock alert: #{error.message}")
  end
end
