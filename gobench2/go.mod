module github.com/arangodb/gobench2

go 1.16

replace github.com/arangodb/go-driver => github.com/arangodb/go-driver v0.0.0-20210906125739-9dc76e43805d

replace github.com/arangodb/go-driver/v2 => github.com/arangodb/go-driver/v2 v2.0.0-20210906125739-9dc76e43805d

require (
	github.com/arangodb/go-driver v0.0.0-20210906111341-1a8bebe2289c // indirect
	github.com/arangodb/go-driver/v2 v2.0.0-20210825071748-9f1169c6a7dc
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
)
