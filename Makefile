build:
	go build

build-static-mac:
	mkdir -p out/ && env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o out/sentlog-Darwin-x86_64

run:
	go build && ./$(notdir $(CURDIR))

clean:
	rm -rf ./sentlog
