all: build/muon build/muoc

build:
	mkdir build

build/muon: build muon/main.go
	go build -v -i -o build/muon muon/*.go

build/muoc: build muoc/main.go
	go build -v -i -o build/muoc muoc/*.go

clean:
	rm -r build
