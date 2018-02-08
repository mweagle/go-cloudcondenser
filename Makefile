.DEFAULT_GOAL=build

GO_LINT := $(GOPATH)/bin/golint
WORK_DIR := ./gocc

################################################################################
# Meta
################################################################################
reset:
	git reset --hard
	git clean -f -d

################################################################################
# Code generation
################################################################################
generate:
	echo "Nothing to generate"

################################################################################
# Hygiene checks
################################################################################

GO_SOURCE_FILES := find . -type f -name '*.go' \
	! -path './vendor/*' \

.PHONY: install_requirements
install_requirements:
	go get -u honnef.co/go/tools/cmd/megacheck
	go get -u honnef.co/go/tools/cmd/gosimple
	go get -u honnef.co/go/tools/cmd/unused
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/fzipp/gocyclo
	go get -u github.com/golang/lint/golint
	go get -u github.com/mjibson/esc

.PHONY: vet
vet: install_requirements
	for file in $(shell $(GO_SOURCE_FILES)); do \
		go tool vet "$${file}" || exit 1 ;\
	done

.PHONY: lint
lint: install_requirements
	for file in $(shell $(GO_SOURCE_FILES)); do \
		$(GO_LINT) "$${file}" || exit 1 ;\
	done

.PHONY: fmt
fmt: install_requirements
	$(GO_SOURCE_FILES) -exec goimports -w {} \;

.PHONY: fmtcheck
fmtcheck:install_requirements
	@ export output="$$($(GO_SOURCE_FILES) -exec goimports -d {} \;)"; \
		test -z "$${output}" || (echo "$${output}" && exit 1)

.PHONY: validate
validate: install_requirements vet lint fmtcheck
	megacheck .

################################################################################
# Travis
################################################################################
travis-depends: install_requirements
	go get -u github.com/golang/dep/...
	dep ensure
	# Move everything in the ./vendor directory to the $(GOPATH)/src directory
	rsync -a --quiet --remove-source-files ./vendor/ $(GOPATH)/src

.PHONY: travis-ci-test
travis-ci-test: travis-depends test build
	go test -v -cover ./...

################################################################################
# ALM commands
################################################################################
.PHONY: ensure-preconditions
ensure-preconditions:
	mkdir -pv $(WORK_DIR)

.PHONY: clean
clean:
	go clean .
	go env

.PHONY: test
test: validate
	go test -v -cover ./...

.PHONY: test-cover
test-cover: ensure-preconditions
	go test -coverprofile=$(WORK_DIR)/cover.out -v .
	go tool cover -html=$(WORK_DIR)/cover.out
	rm $(WORK_DIR)/cover.out
	open $(WORK_DIR)/cover.html

.PHONY: build
build: validate test
	go build ./cmd/main.go
	@echo "Build complete"
