class CreateAllocations < ActiveRecord::Migration[6.1]
  def change
    create_table :allocations, id: false do |t|
      t.string :category, null: false
      t.integer :amount, null: false
    end

    execute "ALTER TABLE allocations ADD PRIMARY KEY (category)"
  end
end
