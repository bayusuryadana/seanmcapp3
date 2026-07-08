module Api
  class BaseController < ActionController::API
    rescue_from ActiveRecord::RecordNotFound, with: :not_found
    rescue_from ActiveRecord::RecordInvalid, with: :validation_failed
    rescue_from ActionController::ParameterMissing, with: :bad_request
    rescue_from AppValidationError, with: :app_validation_failed

    private

    def authenticate_wallet!
      return if JwtToken.valid?(request.authorization.to_s)

      render json: { error: "Invalid token" }, status: :unauthorized
    end

    def render_data(data)
      render json: { data: data }
    end

    def not_found
      render json: { error: "not found" }, status: :not_found
    end

    def validation_failed(error)
      render json: { error: error.record.errors.full_messages.to_sentence }, status: :bad_request
    end

    def app_validation_failed(error)
      render json: { error: error.message }, status: :bad_request
    end

    def bad_request
      render json: { error: "Invalid JSON" }, status: :bad_request
    end
  end
end
