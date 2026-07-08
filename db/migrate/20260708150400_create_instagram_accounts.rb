class CreateInstagramAccounts < ActiveRecord::Migration[6.1]
  def change
    create_table :instagram_accounts, id: false do |t|
      t.string :username, null: false
      t.text :last_shortcodes
    end

    execute "ALTER TABLE instagram_accounts ADD PRIMARY KEY (username)"
  end
end
