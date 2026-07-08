class FrontendController < ApplicationController
  BUILD_PATH = Rails.root.join("ui", ".build").freeze

  def index
    send_file BUILD_PATH.join("index.html"), type: "text/html", disposition: "inline"
  end

  def static
    requested = BUILD_PATH.join("static", params[:path]).cleanpath
    return head :not_found unless requested.to_s.start_with?(BUILD_PATH.join("static").to_s)
    return head :not_found unless File.file?(requested)

    send_file requested, disposition: "inline"
  end
end
