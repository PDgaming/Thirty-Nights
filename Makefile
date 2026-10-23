BINARY := bin/thirty-nights
CMD := ./src

.PHONY: all run build clean fmt vet

all: build

run:
	go run $(CMD)

build:
	mkdir -p bin
	go build -o $(BINARY) $(CMD)

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin