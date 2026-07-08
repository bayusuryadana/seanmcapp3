class Stock < ApplicationRecord
  self.primary_key = :name

  validates :name, presence: true
  validates :best_price, :fair_price, numericality: { greater_than: 0 }

  def dashboard_json
    {
      name: name,
      best_price: best_price,
      current_price: current_price,
      fair_price: fair_price,
      status: status,
      buy_price: buy_price,
      lot: lot
    }.compact
  end
end
