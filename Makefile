build:
	go build

run:
	go build && ./$(notdir $(CURDIR))
