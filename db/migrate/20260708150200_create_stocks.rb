class CreateStocks < ActiveRecord::Migration[6.1]
  def change
    create_table :stocks, id: false do |t|
      t.string :name, null: false
      t.bigint :best_price, null: false
      t.bigint :current_price
      t.bigint :fair_price, null: false
      t.boolean :status, null: false, default: false
      t.bigint :buy_price
      t.bigint :lot
    end

    execute "ALTER TABLE stocks ADD PRIMARY KEY (name)"
  end
end
