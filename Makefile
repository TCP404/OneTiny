GOBUILD=go build
BINARY_NAME=OneTiny

MAC_EXE     :=./exe/$(BINARY_NAME)_mac
LINUX_EXE   :=./exe/$(BINARY_NAME)
WINDOWS_EXE :=./exe/$(BINARY_NAME).exe

BUILD_MAC     := CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 $(GOBUILD) -o $(MAC_EXE)     main.go
BUILD_LINUX   := CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 $(GOBUILD) -o $(LINUX_EXE)   main.go 
BUILD_WINDOWS := CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(WINDOWS_EXE) main.go

COMPRESS_MAC     := upx --best $(MAC_EXE)
COMPRESS_LINUX   := upx --best $(LINUX_EXE)
COMPRESS_WINDOWS := upx --best $(WINDOWS_EXE)

start: clean build compress

build:
	mkdir -p ./exe
	$(BUILD_MAC)
	$(BUILD_LINUX)
	$(BUILD_WINDOWS)

compress:
	$(COMPRESS_MAC)
	$(COMPRESS_LINUX)
	$(COMPRESS_WINDOWS)

.PHONY: clean
clean:
	rm -rf ./exe/*

