class InstagramAccount < ApplicationRecord
  self.primary_key = :username

  validates :username, presence: true
end
