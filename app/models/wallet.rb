class Wallet < ApplicationRecord
  validates :date, :name, :category, :currency, :amount, :account, presence: true

  def dashboard_json
    {
      id: id,
      date: date,
      name: name,
      category: category,
      currency: currency,
      amount: amount,
      done: done,
      account: account
    }
  end
end
