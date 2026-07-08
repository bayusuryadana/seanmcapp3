server_command = File.basename($PROGRAM_NAME) == "puma" || ARGV.first == "server" || ENV["RAILS_START_SCHEDULER"] == "true"
ActiveJob::Base.queue_adapter = :async if server_command
