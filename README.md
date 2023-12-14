# Auth
[![Go Reference](https://pkg.go.dev/badge/github.com/cristosal/auth.svg)](https://pkg.go.dev/github.com/cristosal/auth)

Easy to use postgres backed authentication service for go. 

## Features
- Authentication
- Users
- Groups
- Permissions
- Sessions
- Rate limiting (with redis)
- Password Resets
- Registration Confirmations

## Installation
`go get -u github.com/cristosal/auth`

## Documentation

View the godoc documentation here

https://pkg.go.dev/github.com/cristosal/auth

## Usage

Create a new service using an existing `*sql.DB`

```go
db, _ := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))

authService := auth.NewService(db)
```


You now have access to the various underlying apis

```go
// users api
authService.Users()

// permissions api 
authService.Permissions()

// groups api
authService.Groups()

// sessions api
authService.Sessions()
```

If you want to use rate limiting, pass in a redis client

```go
rcl := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

limiter := auth.NewRedisRateLimiter(rcl)
```


