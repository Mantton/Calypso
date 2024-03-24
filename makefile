.PHONY: run
run:
	@clear
	@go build -tags=llvm16 -o ./bin/calypso ./cmd/calypso.go 
	@./bin/calypso ./dev/dev.test.cly 