build:
	go build -o cmd/gitlab-flagman cmd/main.go

test:
	go test ./...

run:
	go run cmd/main.go

run_p:
	go run cmd/main.go -flagsFile=feature_flags.yaml -gitLabBase=https://gitlab.com/api/v4 -gitLabToken=glpat-sYyrPx4mm5oHcfyd__zD -gitLabProjectID=66045069