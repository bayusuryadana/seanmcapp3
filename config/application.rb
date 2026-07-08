require_relative "boot"

require "logger"
require "net/http"
require "json"
require "set"
require "active_model/railtie"
require "active_job/railtie"
require "active_record/railtie"
require "action_controller/railtie"
require "action_view/railtie"
require "rails/test_unit/railtie"

Bundler.require(*Rails.groups)

module Seanmcapp
  class Application < Rails::Application
    config.load_defaults 6.1

    config.time_zone = "Asia/Jakarta"
    config.active_job.queue_adapter = :inline
    config.public_file_server.enabled = true
    config.secret_key_base = ENV["SECRET_KEY_BASE"].presence || ENV["APPS_SECRET_KEY"]

    config.middleware.insert_before 0, Rack::Cors do
      allow do
        origins "http://localhost:5173", "http://localhost:8080", "https://seanmcapp.herokuapp.com"
        resource "/api/*",
                 headers: %w[Origin Content-Type Authorization],
                 methods: %i[get post put patch delete options],
                 credentials: true,
                 expose: ["Content-Length"],
                 max_age: 12.hours.to_i
      end
    end
  end
end
