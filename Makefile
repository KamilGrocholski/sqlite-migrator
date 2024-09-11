_run-default:
	go run cmd/migrate/main.go

build:
	go build -o migrate cmd/migrate/main.go
