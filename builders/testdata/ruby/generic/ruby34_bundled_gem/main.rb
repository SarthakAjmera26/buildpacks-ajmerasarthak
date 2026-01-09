require "csv"
require "webrick"

port = ENV["PORT"]&.to_i || 8080

server = WEBrick::HTTPServer.new(Port: port)

server.mount_proc "/" do |req, res|
  res.body = "Hello, World!"
end

server.mount_proc "/csv" do |req, res|
  csv_string = "a,b,c\n1,2,3"
  csv = CSV.parse(csv_string)
  res.body = csv[1][2]
end

server.start

