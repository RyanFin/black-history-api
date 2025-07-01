run:
	go run main.go

build:
	GOOS=linux GOARCH=amd64 go build -o main main.go

logs:
	heroku logs --tail

.PHONY: run build logs