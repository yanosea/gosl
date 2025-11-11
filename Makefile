# shows help message defaultly
.DEFAULT_GOAL := help

#
# build
#
.PHONY: build.linux build.darwin build.windows

# build binary for linux amd64
build.linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/gosl-linux-amd64 ./cmd/gosl

# build binaries for darwin amd64/arm64
build.darwin:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/gosl-darwin-amd64 ./cmd/gosl
	GOOS=darwin GOARCH=arm64 go build -o ./bin/gosl-darwin-arm64 ./cmd/gosl

# build binary for windows amd64
build.windows:
	GOOS=windows GOARCH=amd64 go build -o ./bin/gosl-windows-amd64.exe ./cmd/gosl

#
# update
#
.PHONY: update.credits update.mocks

# update `./CREDITS`
update.credits:
	gocredits -skip-missing . > ./CREDITS

# update mocks
update.mocks:
	mockgen -source=internal/app/port/slack_repository.go -destination=internal/app/port/mock/mock_slack_repository.go -package=mock
	mockgen -source=internal/app/port/config_repository.go -destination=internal/app/port/mock/mock_config_repository.go -package=mock
	mockgen -source=internal/app/port/cache_repository.go -destination=internal/app/port/mock/mock_cache_repository.go -package=mock

#
# test
#
.PHONY: test

# execute all tests with coverage
test:
	@set -e; \
	if [ -f "./test.run" ]; then \
		echo "test already running"; \
		exit 1; \
	fi; \
	touch test.run; \
	go test -v -p 1 ./... -cover -coverprofile=./cover.out; \
	grep -v -E "(_mock\.go|/mock/)" ./cover.out > ./cover.out.tmp && mv ./cover.out.tmp ./cover.out; \
	go tool cover -html=./cover.out -o ./docs/coverage.html; \
	rm ./cover.out; \
	if [ -f "./test.run" ]; then \
		rm ./test.run; \
	fi

#
# clean
#
.PHONY: clean

# remove build artifacts
clean:
	@rm -rf ./bin
	@rm -f ./cover.out
	@rm -f ./test.run

#
# help
#
.PHONY: help

help:
	@echo ""
	@echo "available targets:"
	@echo ""
	@echo "  [build]"
	@echo "    build.linux     - build binary for linux amd64"
	@echo "    build.darwin    - build binaries for darwin amd64/arm64"
	@echo "    build.windows   - build binary for windows amd64"
	@echo ""
	@echo "  [update]"
	@echo "    update.credits  - update ./CREDITS file"
	@echo "    update.mocks    - update all mocks"
	@echo ""
	@echo "  [test]"
	@echo "    test            - execute all tests with coverage"
	@echo ""
	@echo "  [clean]"
	@echo "    clean           - remove build artifacts"
	@echo ""
