
test-payload:
	curl -XPOST -H 'X-Hook-Secret:test' \
		-d "{\"time\": \"$(date +"%Y-%m-%dT%H:%M:%S")\", \"content\": \"Hello World\"}" \
		http://localhost:8080/v1alpha1/webhooks/example

install-k6:
	@if ! which k6 > /dev/null; then \
		echo "Installing k6..." \
		sudo gpg -k; \
		sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69; \
		echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list; \
		sudo apt-get update; \
		sudo apt-get install k6; \
		echo "k6 installed successfully"; \
	else \
		echo "k6 is already installed"; \
	fi

build:
	@echo "Building webhooked..."
	@GOOS=linux GOARCH=amd64 go build -o ./bin/webhooked ./main.go

tests: test-units test-integrations

test-units:
	@echo "Running unit tests..."
	@export WH_DEBUG=true
	@go test ./... -coverprofile coverage.out -covermode count
	@go tool cover -func coverage.out

run-integration: build
	@./bin/webhooked --config ./tests/integrations/webhooked_config.integration.yaml serve

test-integrations: install-k6
	@echo "Running integration tests..."
	
	@if ! pgrep -f "./bin/webhooked" > /dev/null; then \
		echo "PID file not found. Please run 'make run-integration' in another terminal."; \
		exit 1; \
	fi

	@echo "Running k6 tests..."
	@k6 run ./tests/integrations/scenarios.js

.PHONY: test-payload install-k6 build run-integration test-integration