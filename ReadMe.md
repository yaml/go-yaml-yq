# go-yaml-yq

A Go binding for yq's expression engine over `go.yaml.in/yaml/v4`
representation nodes.

Status: prototype.

```bash
go get github.com/yaml/go-yaml-yq
```

This module is tested with Go 1.23. The repository Makefiles install Go
1.23.12 locally through Makes, so a system Go installation is not required for
development.

## Purpose

`go-yaml-yq` lets Go programs run yq expressions directly against
`*yaml.Node` values from `go.yaml.in/yaml/v4`.

The public API is intentionally small:

```go
func Nodes(expr string, nodes ...*yaml.Node) ([]*yaml.Node, error)
func Node(expr string, nodes ...*yaml.Node) (*yaml.Node, error)

func RenderYAML(node *yaml.Node, opts ...RenderOption) ([]byte, error)
func RenderJSON(node *yaml.Node, opts ...RenderOption) ([]byte, error)
func WriteYAML(w io.Writer, node *yaml.Node, opts ...RenderOption) error
func WriteJSON(w io.Writer, node *yaml.Node, opts ...RenderOption) error

func DumpNode(node *yaml.Node) ([]byte, error)
func DumpNodeLong(node *yaml.Node) ([]byte, error)
func WriteNodeDump(w io.Writer, node *yaml.Node) error
func WriteNodeDumpLong(w io.Writer, node *yaml.Node) error
```

Use it when you want yq's query and expression language from Go code, while
keeping the YAML representation as go-yaml v4 nodes.

## Contract

`Node` and `Nodes` are pure.

They never mutate input nodes. Every returned node is a fresh detached copy,
not an interior pointer into the source document.

```go
updated, err := yq.Node(".image = \"example/app:2.0\"", doc)
// doc is unchanged.
// updated is a new YAML tree.
```

For live in-place structural operations, use the companion
`github.com/yaml/go-yaml-dom` package. `go-yaml-yq` and `go-yaml-dom` compose
only through `*yaml.Node`; neither imports the other.

While `go.yaml.in/yaml/v4` is pre-1.0, keep this module, `go-yaml-dom`, and
consumers using both on the same `go.yaml.in/yaml/v4` version. Otherwise
`*yaml.Node` can resolve to distinct Go types.

## API

### Nodes

```go
nodes, err := yq.Nodes(".items[]", doc)
```

`Nodes` evaluates a yq expression and returns the full result stream.

The first input node binds to both `.` and `$1`. Additional input nodes bind to
`$2`, `$3`, and so on:

```go
merged, err := yq.Node("$1 * $2", base, overlay)
```

Errors are returned for malformed expressions, nil input nodes, conversion
failures, and evaluation failures.

### Node

```go
node, err := yq.Node(".metadata.name", doc)
```

`Node` is the strict single-result form. It returns an error unless the
expression yields exactly one result.

Use `Nodes` for expressions that may return zero, one, or many results.

### Render YAML And JSON

```go
out, err := yq.RenderYAML(doc)
out, err := yq.RenderJSON(doc)
```

The render helpers use yq's own YAML and JSON encoders. They are useful when
you want to see how yq would print a node graph, rather than using go-yaml's
`yaml.Marshal` directly.

Color is enabled by default:

```go
out, err := yq.RenderYAML(doc, yq.WithColor(false))
```

Options:

```go
yq.WithColor(false)
yq.WithIndent(4)
yq.WithUnwrapScalar(false)
```

The colorizer is dependency-free and internal to this module. It targets useful
terminal coloring for common YAML and JSON output; it is not a promise of exact
byte-for-byte color parity with the yq CLI.

### Dump Node Structure

```go
out, err := yq.DumpNode(doc)
out, err := yq.DumpNodeLong(doc)
```

`DumpNode` returns a compact YAML description of the `yaml.Node` graph, modeled
after `go-yaml -n`.

`DumpNodeLong` returns a profuse YAML description, modeled after `go-yaml -N`,
including node kind, style, tag, comments, scalar text, and content.

These functions do not render the represented YAML value. They print the node
structure itself for inspection and debugging.

## Loading And Printing YAML

Most programs load a YAML document into a document node, then pass
`doc.Content[0]` to this package:

```go
func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}
```

Print result nodes with go-yaml:

```go
out, err := yaml.Marshal(node)
if err != nil {
	return err
}
fmt.Print(string(out))
```

Or render with yq's encoder:

```go
out, err := yq.RenderYAML(node, yq.WithColor(false))
if err != nil {
	return err
}
fmt.Print(string(out))
```

## Expressions

Queries:

```go
name, err := yq.Node(".metadata.name", doc)
items, err := yq.Nodes(".items[]", doc)
keys, err := yq.Nodes(".data | keys | .[]", doc)
```

Updates:

```go
updated, err := yq.Node(".spec.replicas = 3", doc)
deleted, err := yq.Node("del(.metadata.annotations)", doc)
```

Multiple input nodes:

```go
updated, err := yq.Node(".spec.template.metadata.labels = $2", deployment, labels)
merged, err := yq.Node("$1 * $2", base, overlay)
```

Keys containing dots or other special characters need yq bracket syntax:

```go
value, err := yq.Node(`.["weird.key"]`, doc)
```

Operator behavior follows yq's documentation:
https://mikefarah.gitbook.io/yq/operators

## Merge

This package can evaluate yq merge expressions:

```go
merged, err := yq.Node("$1 * $2", base, overlay)
```

That is useful when you intentionally want yq expression semantics.

For the supported whole-node structural merge API, use `go-yaml-dom`:

```go
copy := dom.Clone(base)
err := dom.Merge(copy, overlay)
```

`go-yaml-dom` merge mutates the destination node in place. `go-yaml-yq` returns
a detached copy.

## Example Programs

The repository has buildable programs under `examples/`. Each example supports:

```bash
make build
make run
make clean
```

The same programs are shown below in full.

### Query A File

Run:

```bash
make -C examples/query run
go run ./examples/query examples/query/sample.yaml '.metadata.labels.app'
```

Program:

```go
package main

import (
	"fmt"
	"os"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.yaml> <yq-expression>\n", os.Args[0])
		os.Exit(2)
	}

	root, err := loadRoot(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	nodes, err := yq.Nodes(os.Args[2], root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for i, node := range nodes {
		if i > 0 {
			fmt.Println("---")
		}
		if err := printYAML(node); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}

func printYAML(node *yaml.Node) error {
	out, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Print(string(out))
	return nil
}
```

### Update A File Copy

Run:

```bash
make -C examples/update run
go run ./examples/update examples/update/sample.yaml '.replicas = 3'
```

Program:

```go
package main

import (
	"fmt"
	"os"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.yaml> <yq-update-expression>\n", os.Args[0])
		os.Exit(2)
	}

	root, err := loadRoot(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	updated, err := yq.Node(os.Args[2], root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := printYAML(updated); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}

func printYAML(node *yaml.Node) error {
	out, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Print(string(out))
	return nil
}
```

### Merge Two Files

Run:

```bash
make -C examples/merge run
go run ./examples/merge examples/merge/base.yaml examples/merge/overlay.yaml
```

Program:

```go
package main

import (
	"fmt"
	"os"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <base.yaml> <overlay.yaml>\n", os.Args[0])
		os.Exit(2)
	}

	base, err := loadRoot(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	overlay, err := loadRoot(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	merged, err := yq.Node("$1 * $2", base, overlay)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := printYAML(merged); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}

func printYAML(node *yaml.Node) error {
	out, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Print(string(out))
	return nil
}
```

### Prompt For Expressions

Run:

```bash
make -C examples/prompt run
go run ./examples/prompt examples/prompt/sample.yaml
```

Program:

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.yaml>\n", os.Args[0])
		os.Exit(2)
	}

	root, err := loadRoot(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := prompt(os.Stdin, os.Stdout, root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func prompt(in io.Reader, out io.Writer, root *yaml.Node) error {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "yq> ")
		if !scanner.Scan() {
			break
		}
		expr := strings.TrimSpace(scanner.Text())
		if expr == "" {
			continue
		}
		if expr == "exit" || expr == "quit" {
			break
		}

		nodes, err := yq.Nodes(expr, root)
		if err != nil {
			fmt.Fprintf(out, "error: %v\n", err)
			continue
		}
		for i, node := range nodes {
			if i > 0 {
				fmt.Fprintln(out, "---")
			}
			if err := writeYAML(out, node); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read expression: %w", err)
	}
	return nil
}

func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}

func writeYAML(out io.Writer, node *yaml.Node) error {
	data, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	_, err = out.Write(data)
	return err
}
```

### Render With Color

Run:

```bash
make -C examples/color run
make -C examples/color run FORMAT=json
go run ./examples/color examples/color/sample.yaml yaml
```

Program:

```go
package main

import (
	"fmt"
	"os"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.yaml> [yaml|json]\n", os.Args[0])
		os.Exit(2)
	}

	format := "yaml"
	if len(os.Args) == 3 {
		format = os.Args[2]
	}

	root, err := loadRoot(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch format {
	case "yaml":
		err = yq.WriteYAML(os.Stdout, root, yq.WithColor(true))
	case "json":
		err = yq.WriteJSON(os.Stdout, root, yq.WithColor(true), yq.WithIndent(2))
	default:
		err = fmt.Errorf("format must be yaml or json, got %q", format)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadRoot(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var doc yaml.Node
	if err := yaml.NewDecoder(file).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty YAML document", path)
	}
	return doc.Content[0], nil
}
```

## Development

This repository uses Makes to install Go locally under `.cache/local`; a system
Go installation is not required. The Makefiles currently use Go 1.23.12.

```bash
make test
make test-examples
make test-all
make vet
make examples
make check
make clean
make shell
```

Per-example commands:

```bash
make -C examples/query run
make -C examples/update run
make -C examples/merge run
make -C examples/prompt run
make -C examples/color run
```

## Release

Create an annotated release tag with:

```bash
make release VERSION=0.1.1
```

`VERSION` is required, must be a semantic version like `0.1.1`, and must not
start with `v`. The release target runs verification first, requires a clean
working tree, and creates tag `v$(VERSION)`.

## CI

GitHub Actions runs tests, hygiene checks, example smoke tests, and CodeQL.
Hygiene includes formatting, `go.mod`/`go.sum`, file lint, spelling, and
forbidden dependency checks. The dependency check keeps pruned yq format adapter
dependencies out of the module graph.

Dependabot is configured for Go modules and GitHub Actions.

## Vendored Engine

The expression engine is vendored from `github.com/mikefarah/yq`; see `Notice`.
The vendored copy is pruned to avoid non-core format adapter dependencies while
keeping yq operators available through the expression engine.
