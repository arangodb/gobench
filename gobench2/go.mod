module github.com/arangodb/gobench2

go 1.23.0

replace github.com/arangodb/go-driver => github.com/arangodb/go-driver v0.0.0-20210906125739-9dc76e43805d

replace github.com/arangodb/go-driver/v2 => github.com/arangodb/go-driver/v2 v2.0.0-20210906125739-9dc76e43805d

require (
	github.com/arangodb/go-driver/v2 v2.0.0-20210825071748-9f1169c6a7dc
	golang.org/x/net v0.38.0
)

require (
	github.com/arangodb/go-velocypack v0.0.0-20200318135517-5af53c29c67e // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.19.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
