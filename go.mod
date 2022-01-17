module github.com/grokloc/grokloc-server

go 1.17

replace (
	github.com/grokloc/grokloc-go/pkg/env => ./pkg/env
	github.com/grokloc/grokloc-go/pkg/models => ./pkg/models
	github.com/grokloc/grokloc-go/pkg/models/admin => ./pkg/models/admin
	github.com/grokloc/grokloc-go/pkg/schemas => ./pkg/schemas
	github.com/grokloc/grokloc-go/pkg/security => ./pkg/security
)

require (
	github.com/google/uuid v1.3.0
	github.com/matthewhartstonge/argon2 v0.1.5
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
