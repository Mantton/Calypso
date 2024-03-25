.PHONY: run
run:
	@clear
	@go build -tags=llvm16 -o ./bin/calypso ./cmd/calypso.go 
	@./bin/calypso build ./dev/src/main.cly ./dev/src/assist.cly