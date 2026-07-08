class JwtToken
  SUBJECT = "wallet-user".freeze

  def self.create(password)
    return "" unless password == AppSettings.fetch!("APPS_PASSWORD")

    JWT.encode(
      {
        sub: SUBJECT,
        iat: Time.now.to_i,
        exp: 12.hours.from_now.to_i
      },
      AppSettings.fetch!("APPS_SECRET_KEY"),
      "HS256"
    )
  end

  def self.valid?(authorization)
    token = authorization.sub(/\ABearer /, "")
    payload, = JWT.decode(token, AppSettings.fetch!("APPS_SECRET_KEY"), true, algorithm: "HS256")
    payload["sub"] == SUBJECT
  rescue JWT::DecodeError
    false
  end
end
