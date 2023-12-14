module github.com/cristosal/auth

go 1.21.4

replace github.com/cristosal/orm => ../orm

require (
	github.com/cristosal/orm v0.0.2-beta
	github.com/go-redis/redis/v7 v7.4.1
	github.com/jackc/pgx/v5 v5.5.0
	golang.org/x/crypto v0.15.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
