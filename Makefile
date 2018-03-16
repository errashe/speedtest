all: client

client:
	@GOOS=windows GOARCH=386 go build -ldflags="-w -s" -o out/client32.exe client.go && upx -9 out/client32.exe
	@GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o out/client64.exe client.go && upx -9 out/client64.exe