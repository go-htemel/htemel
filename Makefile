.PHONY: generate test bench lint check-generated tidy

generate:
	go generate ./...

test:
	go test -race ./...

bench:
	go test -run '^$$' -bench . -benchmem ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; falling back to go vet"; \
		go vet ./...; \
	fi

check-generated: generate
	@if [ -n "$$(git status --porcelain html/)" ]; then \
		echo "html/ is out of date; run 'make generate' and commit the result"; \
		git status --porcelain html/; \
		exit 1; \
	fi

tidy:
	go mod tidy
