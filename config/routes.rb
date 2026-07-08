Rails.application.routes.draw do
  namespace :api do
    post "webhook", to: "webhooks#create"

    namespace :wallet do
      post "login", to: "sessions#create"
      get "dashboard", to: "dashboard#show"
      post "create", to: "wallets#create"
      post "update", to: "wallets#update"
      delete "delete/:id", to: "wallets#destroy"
    end

    namespace :stock do
      post "getAll", to: "stocks#index"
      post "refresh", to: "stocks#refresh"
      post "create", to: "stocks#create"
      post "update", to: "stocks#update"
      delete "delete/:id", to: "stocks#destroy"
    end

    namespace :instagram do
      get "trigger", to: "triggers#create"
    end
  end

  get "/static/*path", to: "frontend#static"
  root "frontend#index"
  get "*path", to: "frontend#index"
end
