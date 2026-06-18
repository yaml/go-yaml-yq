R := https://github.com/makeplus/makes
M ?= $(or $(MAKES_REPO_DIR),.cache/makes)
$(shell [ -d '$(M)' ] || git clone -q $(R) '$(M)')

include $(M)/init.mk

GO-VERSION := 1.23.12
GH-TOKEN-FILE := $(HOME)/.github-token-yaml
include $(M)/go.mk
include $(M)/clean.mk

include $(M)/gh.mk

RELEASE-VERSION := $(patsubst v%,%,$(VERSION))
RELEASE-TAG := v$(RELEASE-VERSION)

check: test vet examples

deps: $(GO)
	@go list -m all

fmt: $(GO)
	gofmt -w doc.go yq.go render.go dump.go yq_test.go render_test.go dump_test.go examples_test.go internal/yqlib/color_print_stub.go examples/query/main.go examples/update/main.go examples/prompt/main.go examples/merge/main.go examples/color/main.go

tidy: $(GO)
	go mod tidy

test: $(GO)
	go test ./...

test-examples:: FORCE
	@$(MAKE) --no-print-directory -C examples run

test-all:: test test-examples

vet: $(GO)
	go vet ./...

examples:: $(GO) FORCE
	go build ./examples/query ./examples/update ./examples/prompt ./examples/merge ./examples/color

clean::
	@$(MAKE) --no-print-directory -C examples clean

verify: fmt tidy vet test

release:
ifndef VERSION
	@$(MAKE) --no-print-directory last-release
else
	@$(MAKE) --no-print-directory release-create VERSION=$(VERSION)
endif

last-release:
	@latest="$$(git tag --list 'v*' --sort=-v:refname | head -1)"; \
	if [[ -n "$$latest" ]]; then \
	  echo "$$latest"; \
	else \
	  echo 'No release tags found'; \
	fi

release-create: release-github

release-check:
ifndef VERSION
	@echo "Set VERSION=x.y.z to use 'make release'"
	@exit 1
endif
	@printf '%s\n' '$(RELEASE-VERSION)' | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$' || \
	  (echo "VERSION must be a semantic version like 0.1.1" >&2; exit 1)

release-remote-check: release-check $(GH)
	@$(GH-CMD) release view '$(RELEASE-TAG)' >/dev/null 2>&1 && \
	  (echo "GitHub release $(RELEASE-TAG) already exists" >&2; exit 1) || true

release-tag: release-check verify release-remote-check
	@git diff --quiet -- . ':!.cache' || \
	  (echo "Working tree has uncommitted changes" >&2; exit 1)
	@git diff --cached --quiet -- . ':!.cache' || \
	  (echo "Index has staged changes" >&2; exit 1)
	@test -z "$$(git status --porcelain --untracked-files=all -- . ':!.cache')" || \
	  (echo "Working tree has untracked files" >&2; exit 1)
	@git rev-parse --verify '$(RELEASE-TAG)' >/dev/null 2>&1 && \
	  (echo "Tag $(RELEASE-TAG) already exists" >&2; exit 1) || true
	git tag -a '$(RELEASE-TAG)' -m 'Release $(RELEASE-TAG)'

release-push: release-tag
	git push origin HEAD
	git push origin '$(RELEASE-TAG)'

release-github: release-push
	$(GH-CMD) release create '$(RELEASE-TAG)' \
	  --title '$(RELEASE-TAG)' \
	  --generate-notes

build-example:: $(GO)
ifndef EXAMPLE
	@echo "Set EXAMPLE=... to use 'make build-example'"
	@exit 1
endif
	go build -o examples/$(EXAMPLE)/$(EXAMPLE) ./examples/$(EXAMPLE)

run-example:: $(GO)
ifndef EXAMPLE
	@echo "Set EXAMPLE=... to use 'make run-example'"
	@exit 1
endif
	@$(MAKE) --no-print-directory -C examples/$(EXAMPLE) run

FORCE:

include $(M)/shell.mk
