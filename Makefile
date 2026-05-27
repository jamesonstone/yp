BIN_DIR := ./bin
BINARY := $(BIN_DIR)/yp
CMD := .

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: install
install:
	go install $(CMD)

.PHONY: fmt
fmt:
	gofmt -w main.go cmd internal

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
