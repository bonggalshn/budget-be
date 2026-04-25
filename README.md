# budget-be

Backend service for budget apps built with Go.

## Prerequisites

- Go 1.23+
- PostgreSQL 14+

## Build

```bash
cd budget-be
go build -o api ./cmd/api
```

## Run

Create a `config.yaml` file (see example below), then:

```bash
./api
```

Or run directly with Go:

```bash
go run ./cmd/api
```

## Configuration

Configuration is loaded from `config.yaml`. Environment variables override config file values.

### config.yaml

```yaml
database:
  host: localhost
  port: "5432"
  name: budget
  user: budget_user
  password: ""

jwt:
  secret: change-me-in-production
  expiry: 24h

server:
  host: 0.0.0.0
  port: "8080"

rateLimit:
  ipRequestsPerMinute: 10
  usernameRequestsPerMinute: 5
  windowMinutes: 15
```

### Configuration Reference

| Config Key | Environment Variable | Description | Default |
|------------|---------------------|-------------|-----------|
| database.host | DB_HOST | PostgreSQL host | localhost |
| database.port | DB_PORT | PostgreSQL port | 5432 |
| database.name | DB_NAME | Database name | budget |
| database.user | DB_USER | Database user | budget_user |
| database.password | DB_PASSWORD | Database password | - |
| jwt.secret | JWT_SECRET | Secret key for JWT signing | change-me-in-production |
| jwt.expiry | JWT_EXPIRY | Token expiration duration | 24h |
| server.host | SERVER_HOST | Server listen address | 0.0.0.0 |
| server.port | SERVER_PORT | Server port | 8080 |