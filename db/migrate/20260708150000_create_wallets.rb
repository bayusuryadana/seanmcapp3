class CreateWallets < ActiveRecord::Migration[6.1]
  def change
    create_table :wallets do |t|
      t.integer :date, null: false
      t.string :name, null: false
      t.string :category, null: false
      t.string :currency, null: false
      t.integer :amount, null: false
      t.boolean :done, null: false, default: false
      t.string :account, null: false
    end
  end
end
