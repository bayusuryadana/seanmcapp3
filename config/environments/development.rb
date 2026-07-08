Rails.application.configure do
  config.cache_classes = false
  config.eager_load = false
  config.consider_all_requests_local = true
  config.server_timing = true
  config.active_record.migration_error = false
  config.active_record.verbose_query_logs = true
  config.log_level = :debug
end
