Rails.application.config.filter_parameters += [
  :password,
  :pass,
  :secret,
  :secret_key,
  :token,
  :session_id,
  :csrf_token
]
