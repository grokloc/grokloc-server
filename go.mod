module github.com/grokloc/grokloc-server

go 1.18

replace (
	github.com/grokloc/grokloc-server/pkg/app => ./pkg/app
	github.com/grokloc/grokloc-server/pkg/app/admin => ./pkg/app/admin
	github.com/grokloc/grokloc-server/pkg/app/admin/org => ./pkg/app/admin/org
	github.com/grokloc/grokloc-server/pkg/app/admin/org/events => ./pkg/app/admin/org/events
	github.com/grokloc/grokloc-server/pkg/app/admin/org/testing => ./pkg/app/admin/org/testing
	github.com/grokloc/grokloc-server/pkg/app/admin/user => ./pkg/app/admin/user
	github.com/grokloc/grokloc-server/pkg/app/admin/user/events => ./pkg/app/admin/user/events
	github.com/grokloc/grokloc-server/pkg/app/admin/user/testing => ./pkg/app/admin/user/testing
	github.com/grokloc/grokloc-server/pkg/app/audit => ./pkg/app/audit
	github.com/grokloc/grokloc-server/pkg/app/audit/testing => ./pkg/app/audit/testing
	github.com/grokloc/grokloc-server/pkg/app/state => ./pkg/app/state
	github.com/grokloc/grokloc-server/pkg/env => ./pkg/env
	github.com/grokloc/grokloc-server/pkg/grokloc => ./pkg/grokloc
	github.com/grokloc/grokloc-server/pkg/models => ./pkg/models
	github.com/grokloc/grokloc-server/pkg/safe => ./pkg/safe
	github.com/grokloc/grokloc-server/pkg/security => ./pkg/security
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi/v5 v5.0.7
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
