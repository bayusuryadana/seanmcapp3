class StockAlertJob < ApplicationJob
  queue_as :default

  def perform
    RunGuard.run("stock refresh") do
      StockAlertService.new.run
    end
  end
end
