require "csv"

CSV.parse("a,b,c") do |row|
  p row
end
