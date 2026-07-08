module Api
  module Instagram
    class TriggersController < Api::BaseController
      def create
        InstagramFetchJob.perform_later
        render json: { data: "Instagram fetch triggered" }
      end
    end
  end
end
