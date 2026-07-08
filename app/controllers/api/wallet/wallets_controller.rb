module Api
  module Wallet
    class WalletsController < Api::BaseController
      before_action :authenticate_wallet!

      def create
        wallet = ::Wallet.create!(wallet_params)
        render_data wallet.id
      end

      def update
        wallet = ::Wallet.find(wallet_params.fetch(:id))
        wallet.update!(wallet_params.except(:id))
        render_data wallet.id
      end

      def destroy
        wallet = ::Wallet.find(params[:id])
        wallet.destroy!
        render_data wallet.id
      end

      private

      def wallet_params
        params.permit(:id, :date, :name, :category, :currency, :amount, :done, :account)
      end
    end
  end
end
