export PATH := $(PWD)/build-dependencies:$(PATH)

.PHONY: all
all: build/muon build/muctl

build:
	mkdir build

build-dependencies:
	mkdir build-dependencies

build-dependencies/genny: build-dependencies
	go build -o build-dependencies github.com/dimchansky/genny

muon/%_list.go: build-dependencies/genny
	go generate ./muon

.PHONY: build/muon
build/muon: build muon/%_list.go
	go build -o build/muon ./cmd/muon

.PHONY: build/muctl
build/muctl: build
	go build -o build/muctl ./cmd/muctl

.PHONY: clean
clean:
	rm -rf build
	rm -f muon/*_list.go
	rm -rf build-dependencies

.PHONY: dev0
dev0: build/muctl muon/%_list.go
	DISPLAY=:0.0 go run ./cmd/muon

.PHONY: dev1
dev1: build/muctl muon/%_list.go
	DISPLAY=:1.0 go run ./cmd/muon

.PHONY: dev0-race
dev0-race: build/muctl muon/%_list.go
	DISPLAY=:0.0 go run -race ./cmd/muon

.PHONY: dev1-race
dev1-race: build/muctl muon/%_list.go
	DISPLAY=:1.0 go run -race ./cmd/muon

.PHONY: install
install: all
	sudo cp -v build/muon /usr/local/bin/muon
	sudo cp -v build/muctl /usr/local/bin/muctl
