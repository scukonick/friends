.PHONY: mocks

deps:
	go mod vendor

mocks:
	rm -rf internal/generated/mocks
	mkdir -p internal/generated/mocks
	mockgen -package mocks -source internal/bus/interface.go -destination internal/generated/mocks/bus.go
	mockgen -package mocks -source internal/dispatcher/interface.go -destination internal/generated/mocks/disp.go
