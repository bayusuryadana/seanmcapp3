class Allocation < ApplicationRecord
  self.primary_key = :category

  validates :category, :amount, presence: true
end
