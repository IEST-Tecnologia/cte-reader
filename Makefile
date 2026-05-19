CHUNK_SIZE ?= 500000

build:
	GOOS=windows GOARCH=amd64 go build -tags prod -ldflags "-H windowsgui -X main.chunkSizeStr=$(CHUNK_SIZE)" -o cte-reader.exe .
