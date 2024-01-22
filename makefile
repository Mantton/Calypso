.PHONY: run repl file

# Define the 'run' target
run:
	@go build ./cmd/calypso.go 
	@./calypso ./dev.test.cly -panic