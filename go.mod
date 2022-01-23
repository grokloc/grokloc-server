module github.com/grokloc/grokloc-server

go 1.17

replace (
	github.com/grokloc/grokloc-server/pkg/app => ./pkg/app
	github.com/grokloc/grokloc-server/pkg/env => ./pkg/env
	github.com/grokloc/grokloc-server/pkg/models => ./pkg/models
	github.com/grokloc/grokloc-server/pkg/models/admin => ./pkg/models/admin
	github.com/grokloc/grokloc-server/pkg/models/admin/org => ./pkg/models/admin/org
	github.com/grokloc/grokloc-server/pkg/models/admin/org/testing => ./pkg/models/admin/org/testing
	github.com/grokloc/grokloc-server/pkg/models/admin/user => ./pkg/models/admin/user
	github.com/grokloc/grokloc-server/pkg/models/admin/user/testing => ./pkg/models/admin/user/testing
	github.com/grokloc/grokloc-server/pkg/schemas => ./pkg/schemas
	github.com/grokloc/grokloc-server/pkg/security => ./pkg/security
	github.com/grokloc/grokloc-server/pkg/state => ./pkg/state
)

require (
	github.com/google/uuid v1.3.0
	github.com/matthewhartstonge/argon2 v0.1.5
	github.com/mattn/go-sqlite3 v1.14.10
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.20.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sys v0.0.0-20210909193231-528a39cd75f3 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
