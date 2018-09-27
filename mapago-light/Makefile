all:
	GOOS=linux GOARCH=amd64 go build mapago-light.go
	GOOS=windows GOARCH=amd64 go build -o mapago-light.exe mapago-light.go

clean:
	rm -rf mapago-light

fmt:
	gofmt -l -w mapago-light.go
