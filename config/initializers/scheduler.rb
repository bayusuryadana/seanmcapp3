server_command = File.basename($PROGRAM_NAME) == "puma" || ARGV.first == "server" || ENV["RAILS_START_SCHEDULER"] == "true"
return unless server_command
return if Rails.env.test? || ENV["DISABLE_SCHEDULER"] == "true"

scheduler = Rufus::Scheduler.singleton

scheduler.cron "0 8 * * *", timezone: "Asia/Jakarta" do
  BirthdayReminderJob.perform_later
end

scheduler.cron "0 9 * * *", timezone: "Asia/Jakarta" do
  NewsDigestJob.perform_later
end

scheduler.cron "0 19 * * *", timezone: "Asia/Jakarta" do
  StockAlertJob.perform_later
end

scheduler.cron "0 10 * * *", timezone: "Asia/Jakarta" do
  InstagramFetchJob.perform_later
end
