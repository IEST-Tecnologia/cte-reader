build:
	GOOS=windows GOARCH=amd64 go build -tags prod -ldflags "-H windowsgui" -o cte-reader.exe .
