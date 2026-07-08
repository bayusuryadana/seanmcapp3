class BirthdayReminderService
  DAY_WORDS = {
    0 => "Today",
    1 => "Tomorrow",
    7 => "Next week"
  }.freeze

  def initialize(telegram_client: TelegramClient.new, personal_chat_id: AppSettings.fetch!("TELEGRAM_PERSONAL_CHAT_ID").to_i)
    @telegram_client = telegram_client
    @personal_chat_id = personal_chat_id
  end

  def run
    now = Time.zone.now
    count = send_for_day(now, 0)
    count += send_for_day(now + 1.day, 1)
    count += send_for_day(now + 7.days, 7)
    Rails.logger.info("[INFO] #{count} people has birthday today")
  end

  private

  def send_for_day(date, day_offset)
    people = Person.where(day: date.day, month: date.month)
    people.each { |person| send_birthday_reminder(person, day_offset) }
    people.size
  end

  def send_birthday_reminder(person, day_offset)
    day_word = DAY_WORDS.fetch(day_offset)
    response = @telegram_client.send_message(@personal_chat_id, "#{day_word} is #{person.name}'s birthday!!")
    Rails.logger.info("Sent birthday message: #{response.inspect}")
  rescue KeyError
    Rails.logger.error("Invalid numOfDays: #{day_offset}")
  rescue StandardError => error
    Rails.logger.error("Failed to send message: #{error.message}")
  end
end
