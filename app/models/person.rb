class Person < ApplicationRecord
  self.table_name = "people"

  validates :name, :day, :month, presence: true
end
