module Api
  module Stock
    class StocksController < Api::BaseController
      before_action :authenticate_wallet!

      def index
        render_data ::Stock.order(:name).map(&:dashboard_json)
      end

      def refresh
        StockRefreshService.new.refresh_prices
        render_data ::Stock.order(:name).map(&:dashboard_json)
      end

      def create
        stock = ::Stock.create!(stock_params)
        render_data stock.name
      end

      def update
        stock = ::Stock.find(stock_params.fetch(:name))
        stock.update!(stock_params.except(:name))
        render_data stock.name
      end

      def destroy
        stock = ::Stock.find(params[:id])
        stock.destroy!
        render_data stock.name
      end

      private

      def stock_params
        params.permit(:name, :best_price, :current_price, :fair_price, :status, :buy_price, :lot)
      end
    end
  end
end
