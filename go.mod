module github.com/mikegin/haversine

go 1.22.4

require github.com/mikegin/gjson v0.0.0
require github.com/mikegin/utils v0.0.0

require (
	github.com/mikegin/match v0.0.0 // indirect
	github.com/mikegin/pretty v0.0.0 // indirect
)

replace github.com/mikegin/gjson => ./gjson

replace github.com/mikegin/match => ./match

replace github.com/mikegin/pretty => ./pretty

replace github.com/mikegin/utils => ./utils
