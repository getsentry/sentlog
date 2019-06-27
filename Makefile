OUTDIR := out

build:
	go build

prepare-dir:
	@mkdir -p ${OUTDIR}

build-static-mac: prepare-dir
	GOOS=darwin GOARCH=amd64 go build -a -o ${OUTDIR}/sentlog-Darwin-x86_64

build-static-linux: prepare-dir
	GOOS=linux GOARCH=amd64 go build -a -o ${OUTDIR}/sentlog-Linux-x86_64

build-static-all: build-static-mac build-static-linux

run:
	go build && ./$(notdir $(CURDIR))

clean:
	rm -rf ./sentlog
