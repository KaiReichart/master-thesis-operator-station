# Makefile for building and running the Go application

APP_NAME=master-thesis-operator-station
GO_FILES=$(wildcard *.go)

.PHONY: all run build clean cross-compile generate

all: generate build

run:
	go tool air

generate:
	go generate ./...

build: generate
	go build -o $(APP_NAME)

clean:
	rm -f $(APP_NAME) $(APP_NAME)-windows.exe $(APP_NAME)-macos
	find . -name "*_templ.go" -delete

cross-compile: generate
	GOOS=windows GOARCH=amd64 go build -o $(APP_NAME)-windows.exe
	GOOS=darwin GOARCH=amd64 go build -o $(APP_NAME)-macos


r:
	R < data_analysis_in_r/analysis.r --no-save