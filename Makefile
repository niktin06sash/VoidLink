include .env
export

BINARY=vlink
CMD_DIR=./cmd/vlink
LOCAL_RUN=./local-env.sh

.PHONY: all build clean run-local status logs ping-test shell-s shell-c unit-test test-cover

all: build

build:
	@echo "==> Building $(BINARY)..."
	go build -o $(BINARY) $(CMD_DIR)

run-local: unit-test build
	@echo "==> Starting local netns environment..."
	chmod +x $(LOCAL_RUN)
	sudo -E $(LOCAL_RUN)

status:
	@echo "==> Displaying status..."
	watch -n 1 "echo '--- SERVER ---' && sudo ./$(BINARY) status server --config $(CFG_S) && echo '\n--- CLIENT ---' && sudo ./$(BINARY) status client --config $(CFG_C)"

logs:
	@echo "==> Following logs..."
	tail -f $(LOG_S) $(LOG_C)

ping-test:
	@echo "==> Pinging Server from Client netns..."
	sudo ip netns exec $(NS_C) ping -c 4 10.1.1.1
	@echo "\n==> Pinging Client from Server netns..."
	sudo ip netns exec $(NS_S) ping -c 4 10.1.1.2

shell-s:
	sudo ip netns exec $(NS_S) bash

shell-c:
	sudo ip netns exec $(NS_C) bash

stop:
	sudo pkill -f "./$(BINARY) up" || true

clean:
	@echo "==> Cleaning up artifacts..."
	rm -f $(BINARY)
	sudo rm -f $(CFG_S) $(CFG_C) $(LOG_S) $(LOG_C)
	sudo rm -f $$(dirname $(CFG_S))/server.key $$(dirname $(CFG_C))/client.key
	sudo rm -f /tmp/vlink-*.sock
	sudo ip netns del $(NS_S) 2>/dev/null || true
	sudo ip netns del $(NS_C) 2>/dev/null || true

unit-test:
	@echo "Running unit-tests with race detector..."
	go test -v -race ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out