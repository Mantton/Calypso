.PHONY: run repl file

# Define the 'run' target
run:
	@go build -o ./bin/calypso ./cmd/calypso.go
	@./bin/calypso ./dev.test.cly -panic
