# Go Syntax Verification Script
# Run this in PowerShell if Go is installed:
# $env:Path += ";C:\Program Files\Go\bin"
# go vet ./...

# File structure check
# Files that should exist:
# budget-be/go.mod
# budget-be/internal/config/config.go
# budget-be/internal/db/pool.go
# budget-be/internal/user/model.go
# budget-be/internal/user/repository.go
# budget-be/internal/session/model.go
# budget-be/internal/session/repository.go
# budget-be/internal/loginattempt/model.go
# budget-be/internal/auth/service.go
# budget-be/migrations/auth/*.sql
# budget-be/tests/*_test.go

# To verify syntax when Go is installed:
# go build -o /dev/null ./budget-be/...
# go test -c ./budget-be/...
# go vet ./budget-be/...
# gofmt -l ./budget-be/...