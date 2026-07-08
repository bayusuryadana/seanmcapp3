class InstagramFetchJob < ApplicationJob
  queue_as :default

  def perform
    RunGuard.run("instagram run") do
      InstagramFetchService.new.run
    end
  end
end
