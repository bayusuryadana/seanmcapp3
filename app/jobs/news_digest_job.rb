class NewsDigestJob < ApplicationJob
  queue_as :default

  def perform
    RunGuard.run("news run") do
      NewsDigestService.new.run
    end
  end
end
