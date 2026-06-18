Contributing to go-yaml-yq
==========================

Thank you for your interest in contributing to go-yaml-yq.

This repository provides yq expression evaluation over `go.yaml.in/yaml/v4`
representation nodes. Keep changes small, well tested, and focused on that yq
expression surface.


## Development

The Makefile bootstraps Makes into `.cache/makes` and installs Go locally, so a
system Go installation is not required.

This module supports Go 1.23. The repository Makefiles use Go 1.23.12.

Useful targets:

```sh
make test        # go test ./...
make vet         # go vet ./...
make verify      # fmt, tidy, vet, test
make examples    # build all example programs
make test-all    # tests plus example smoke runs
make clean       # remove example binaries
make deps        # print the module graph
```

Run a single example:

```sh
make -C examples/query run
```


## Coding Conventions

- Keep the public API small and node-oriented.
- Preserve the pure contract: `Node` and `Nodes` never mutate inputs, and
  returned nodes are detached copies.
- Keep `go-yaml-yq` and `go-yaml-dom` composed only through `*yaml.Node`;
  neither package should import the other.
- Keep the dependency graph pruned so yq format adapter dependencies do not
  return.
- Use `make verify` before sending changes.
- Add or update tests for behavior changes.
- Update `ReadMe.md` and example READMEs when public behavior changes.


## Release Tags

Create a release with:

```sh
make release VERSION=0.1.1
```

`VERSION` is required and may be written as `0.1.1` or `v0.1.1`. The release
target normalizes both forms to the tag `v0.1.1`, runs verification, requires a
clean working tree, pushes the branch and tag to `origin`, and creates a GitHub
release with generated notes.

Run `make release` without `VERSION` to print the latest local release tag
without changing anything.


## Commit Conventions

- Avoid merge commits.
- Commit subject line should:
  - Start with a capital letter.
  - Not end with a period.
  - Be between 20 and 50 characters.
  - Not use conventional-commit prefixes such as `fix:` or `feat:`.
- Separate subject and body with a blank line.


## Pull Requests

1. Create a focused branch.
1. Make the smallest practical change.
1. Add tests and documentation when behavior changes.
1. Run `make verify` and `make test-all`.
1. Submit a pull request.
