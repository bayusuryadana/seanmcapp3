class RunGuard
  @mutex = Mutex.new
  @running = {}

  class << self
    def run(name)
      acquired = false
      @mutex.synchronize do
        unless @running[name]
          @running[name] = true
          acquired = true
        end
      end

      unless acquired
        Rails.logger.info("[INFO] #{name} already in progress, skipping")
        return
      end

      yield
    ensure
      @mutex.synchronize { @running[name] = false } if acquired
    end
  end
end
