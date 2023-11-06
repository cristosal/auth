# Auth
Easy to use Authentication library for go. 

**Note: postgres is the only supported database**

## Features
- Authentication
- Users
- Groups
- Permissions
- Sessions
- Rate limiting (with redis)
- Password Resets
- Registration Confirmations

## Usage

Create a new service using an existing pgx connection

```go

db, _ := pgx.Connect(context.Background(), os.Getenv("CONNECTION_STRING"))

authService := auth.NewPgxService(db)
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