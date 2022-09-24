module github.com/omerhorev/gobash

go 1.18

require github.com/stretchr/testify v1.8.0

replace github.com/omerhorev/gobash/mocks => ./mock

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)