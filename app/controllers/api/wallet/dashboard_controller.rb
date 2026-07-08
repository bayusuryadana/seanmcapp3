module Api
  module Wallet
    class DashboardController < Api::BaseController
      before_action :authenticate_wallet!

      def show
        render_data WalletDashboard.new(params[:date].to_i).as_json
      end
    end
  end
end
