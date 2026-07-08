class BirthdayReminderJob < ApplicationJob
  queue_as :default

  def perform
    RunGuard.run("birthday run") do
      BirthdayReminderService.new.run
    end
  end
end
