module Api
  class WebhooksController < BaseController
    def create
      render json: { status: "ok" }
    end
  end
end
