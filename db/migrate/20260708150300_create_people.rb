class CreatePeople < ActiveRecord::Migration[6.1]
  def change
    create_table :people do |t|
      t.string :name, null: false
      t.integer :day, null: false
      t.integer :month, null: false
    end

    add_index :people, %i[day month]
  end
end
