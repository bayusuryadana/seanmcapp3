module Api
  module Wallet
    class SessionsController < Api::BaseController
      def create
        token = JwtToken.create(params.require(:password))
        return render plain: "Invalid password", status: :unauthorized if token.blank?

        render plain: token
      end
    end
  end
end
