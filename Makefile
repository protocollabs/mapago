all:
	GOOS=linux GOARCH=amd64 go install cmd/server/server.go