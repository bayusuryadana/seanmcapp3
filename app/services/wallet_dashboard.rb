class WalletDashboard
  EXPENSE_CATEGORIES = ["Daily", "Rent", "Travel", "Fashion", "IT Stuff", "Misc", "Wellness", "Funding"].freeze

  def initialize(date)
    @date = date
    @wallets = Wallet.all.to_a
  end

  def as_json
    {
      chart: { balance: balance_history },
      allocations: allocations,
      savings: {
        dbs: total_amount("DBS", nil),
        bca: total_amount("BCA", nil)
      },
      planned: {
        sgd: total_amount("DBS", @date),
        idr: total_amount("BCA", @date)
      },
      detail: @wallets.select { |wallet| wallet.date == @date }.map(&:dashboard_json)
    }
  end

  private

  def balance_history
    sums = @wallets.each_with_object(Hash.new(0)) do |wallet, result|
      result[wallet.date] += wallet.amount if wallet.account == "DBS" && wallet.date <= @date
    end

    total = 0
    cumulative = sums.sort.map do |date, sum|
      total += sum
      { date: date, sum: total }
    end

    cumulative.last(6).sort_by { |row| -row[:date] }
  end

  def allocations
    ytd_allocations = Allocation.all.index_by(&:category)
    ytd_expenses = yearly_expenses

    EXPENSE_CATEGORIES.map do |category|
      {
        name: category,
        expense: ytd_expenses[category] || 0,
        alloc: ytd_allocations[category]&.amount || 0
      }
    end
  end

  def yearly_expenses
    year = @date / 100

    @wallets.each_with_object(Hash.new(0)) do |wallet, expenses|
      next unless wallet.done && wallet.date / 100 == year
      next unless EXPENSE_CATEGORIES.include?(wallet.category)

      case wallet.account
      when "DBS"
        expenses[wallet.category] -= wallet.amount
      when "BCA"
        expenses[wallet.category] -= wallet.amount / 12_700
      end
    end
  end

  def total_amount(account, date)
    @wallets.sum do |wallet|
      if date
        wallet.account == account && date >= wallet.date ? wallet.amount : 0
      else
        wallet.account == account && wallet.done ? wallet.amount : 0
      end
    end
  end
end
