module github.com/neennera/fishertimer/services/admin

go 1.25.0

require (
	github.com/lib/pq v1.10.9
	github.com/neennera/fishertimer/pkg/discovery v0.0.0
)

replace github.com/neennera/fishertimer/pkg/discovery => ../../pkg/discovery
