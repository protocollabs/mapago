all:
	GOOS=linux GOARCH=amd64 go install cmd/server/server.go
	GOOS=linux GOARCH=amd64 go install cmd/client/client.go