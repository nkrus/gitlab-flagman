build:
	go build -o cmd/gitlab-flagman cmd/main.go

test:
	go test ./...

run:
	go run cmd/main.go