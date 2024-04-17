.PHONY: run
run:
	@clear
	@echo "building folder $(FOLDER)"
	@go build -tags=llvm16 -o ./bin/calypso ./cmd/calypso.go 
	@./bin/calypso build ./dev/$(FOLDER)


.PHONY: run A B C
A B C:
	@$(MAKE) run FOLDER=$@