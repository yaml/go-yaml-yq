# go-yaml-yq — Prototype Implementation Plan

> A straightforward Go binding for yq's expression engine, operating on go-yaml v4
> representation nodes (`*yaml.Node`). Run any yq expression — ~65 operators —
> directly from Go code. Its entire public surface is `Node` and `Nodes`.
>
> Companion to **go-yaml-dom** (native, dependency-free node manipulation, incl.
> **merge**). The two compose through the standard `*yaml.Node` type; neither imports
> the other.

This document is written to be handed to a CLI coding agent.

> **Merge lives in go-yaml-dom, not here.** Merge is whole-node structural and needs
> no path/expression, so it is implemented natively in the DOM module (where it stays
> dependency-free and ports to the rest of the YAML family). This module keeps the
> path/expression-based evaluation only. There are no typed `Set`/`Delete`
> conveniences — those are just expressions: `yq.Node(".a = $2", doc, v)` and
> `yq.Node("del(.a)", doc)` both return the modified copy.

> **Naming note:** public functions drop the `Yq` prefix to avoid stutter with the
> package name (`yq.Node`, not `yq.YqNode`; cf. `http.Server`). Reverting to
> `YqNodes`/`YqNode` is a single find-replace, at the cost of a `revive` lint.

---

## 0. Objective and strategy

Ship a working prototype quickly by **vendoring** the minimal yq evaluation engine
behind a small, pure wrapper. Later, circle back with yq's maintainer to replace the
vendored copy with a shared upstream dependency; the vendoring is structured to make
that swap a one-line change.

**Why vendor rather than depend on `yqlib`:** published `yqlib` drags in HCL, TOML,
Lua, INI, properties, go-cty, and a second YAML parser (goccy) via its single-package
format adapters. We want a minimal graph. yq already ships build-tag stubs
(`no_*.go`) for every non-core format, which we exploit to strip those deps (Phase 3).

---

## 1. Locked design decisions

- **Module path:** `github.com/yaml/go-yaml-yq` (placeholder).
- **Package name:** `yq`. Call sites: `yq.Node(...)`, `yq.Nodes(...)`.
- **YAML library:** `go.yaml.in/yaml/v4`, pinned to **`v4.0.0-rc.5`** — same version
  as the vendored engine and as go-yaml-dom (§1.3).
- **Vendored engine location:** `internal/yqlib/` (matches the `package yqlib`
  declaration so copied files stay byte-for-byte unmodified).
- **Naming convention:** node-resolving functions are a singular/plural pair with a
  `…Node`/`…Nodes` suffix (matches `dom.FindNode`/`FindNodes`).

### 1.1 Contract — this module is PURE (copy semantics)

Every function runs the yq engine, which deep-copies inputs into its own
`CandidateNode`, mutates internally, and marshals **fresh** `yaml.Node`s out.
Therefore inputs are never mutated and results are orphan copies, disconnected from
any source document. In-place mutation is **go-yaml-dom's** job. Note explicitly:
`dom.Update(yq.Nodes(expr, doc), fn)` mutates copies and is a no-op on the
source.

### 1.2 MVP public API

- `Nodes(expr string, nodes ...*yaml.Node) ([]*yaml.Node, error)` — full result
  stream. node[0] binds to `.` and `$1`; node[i] binds to `$(i+1)`.
- `Node(expr string, nodes ...*yaml.Node) (*yaml.Node, error)` — strict: error unless
  exactly one result.

That's the whole API. Setting and deleting are expressions, not methods:
`yq.Node(".a.b = $2", doc, value)` returns a copy with the path set; `yq.Node("del(.a.b)",
doc)` returns a copy with the path removed. Keys with dots/special chars need yq's
bracket form in the expression, e.g. `.["weird.key"]`. (Merge is **not** here; use
`dom.Merge`.)

### 1.3 Interop with go-yaml-dom

- Interop surface is exclusively `*yaml.Node`. Compose: build/transform with
  `yq.Node`, overlay with `dom.Merge`; or locate live targets with `dom.FindNodes`,
  compute replacements with `yq.Node`.
- **Cross-module version coordination (critical):** go-yaml-yq, go-yaml-dom, and any
  consumer using both must resolve to the **same** `go.yaml.in/yaml/v4` version, or
  `*yaml.Node` is two distinct types. While yaml/v4 is pre-1.0, keep `require`
  versions identical and bump in lockstep.
- Neither module imports the other.

---

## 2. Repository scaffold

```
go-yaml-yq/
├── go.mod                  # module …/go-yaml-yq; require go.yaml.in/yaml/v4 v4.0.0-rc.5
├── LICENSE
├── NOTICE                  # attribution for vendored yq code (§5)
├── README.md               # §6
├── yq.go                   # wrapper: Nodes, Node (§4)
├── yq_test.go              # tests incl. fidelity round-trip and purity (§7)
├── examples_test.go
└── internal/
    └── yqlib/              # vendored, pruned yq engine (§3); files unmodified
```

```bash
go mod init github.com/yaml/go-yaml-yq
go get go.yaml.in/yaml/v4@v4.0.0-rc.5
```

**Verify:** `go mod download` succeeds.

---

## 3. Vendor the minimal yq engine

### 3.1 Copy the engine package

1. Clone yq at a commit whose `go.mod` pins `go.yaml.in/yaml/v4 v4.0.0-rc.5` (read
   yq's `go.mod`; if it moved to a newer yaml/v4, pin **this** project to match).
2. Copy every non-test `.go` file from yq's `pkg/yqlib/` into `internal/yqlib/`. **Do
   not** copy `*_test.go`, `doc/`, or anything under yq's `cmd/`.
3. Do **not** edit copied files. Directory named `yqlib` keeps `package yqlib` valid.
   Pruning = deleting whole files only — keeps the upstream diff mechanical and the
   future dependency swap one line.

**Verify:** `go build ./internal/...` compiles (full heavy dep set — expected).
`go mod tidy`.

### 3.2 Strip format adapters

For each format: delete its real decoder/encoder file(s); make its `no_*.go` stub
unconditional by removing the `//go:build yq_no<fmt>` line. The stub defines the
constructor symbols `format.go` references (as error stubs), so the registry compiles.

| Format | Delete | Keep (unconditionalize) | Dep removed |
|---|---|---|---|
| HCL  | `decoder_hcl.go`, `encoder_hcl.go`   | `no_hcl.go`   | `hashicorp/hcl/v2` + `zclconf/go-cty` (+ tails) |
| TOML | `decoder_toml.go`, `encoder_toml.go` | `no_toml.go`  | `pelletier/go-toml/v2` |
| Lua  | `decoder_lua.go`, `encoder_lua.go`   | `no_lua.go`   | `yuin/gopher-lua` |
| INI  | `decoder_ini.go`, `encoder_ini.go`   | `no_ini.go`   | `go-ini/ini` |
| Props| `decoder_properties.go`, `encoder_properties.go` | `no_props.go` | `magiconair/properties` |

**Also drop goccy:** delete `color_print.go`, `decoder_goccy_yaml.go`,
`candidate_node_goccy_yaml.go`; stub/fix any dangling reference.

**KEEP:** YAML decoder/encoder, JSON (`goccy/go-json`), CSV/TSV and XML (stdlib),
base64, uri, sh, shell, kyaml, participle lexer, `orderedmap`, `utfbom`, `x/text`.
**Keep all 65 operators** — wired into the lexer rule table; do not hand-prune.

### 3.3 Iterate to a clean graph

```bash
go build ./...
go mod tidy
go list -m all   # hcl, go-cty, gopher-lua, go-toml, go-ini, magiconair, goccy/go-yaml ABSENT
```

`go mod why <module>` traces stragglers.

**Verify (Phase 3 acceptance):** build passes; dropped deps absent; remaining graph ≈
`go.yaml.in/yaml/v4`, `participle/v2`, `orderedmap`, `goccy/go-json`, `utfbom`,
`x/text` (plus small tails).

---

## 4. The wrapper (`yq.go`)

Verified surface: `yqlib.InitExpressionParser()`,
`yqlib.ExpressionParser.ParseExpression(string) (*ExpressionNode, error)`,
`yqlib.NewDataTreeNavigator()` with
`GetMatchingNodes(Context, *ExpressionNode) (Context, error)`; `yqlib.Context`
(`MatchingNodes *list.List`, `SetVariable(name string, value *list.List)`); on
`*yqlib.CandidateNode`: `UnmarshalYAML(*yaml.Node, map[string]*CandidateNode) error`
and `MarshalYAML() (*yaml.Node, error)`. Variable lexer rule `\$[a-zA-Z_\-0-9]+`,
leading `$` stripped, so `$1` binds via `SetVariable("1", …)`.

```go
// Package yq is a Go binding for yq's expression engine over go-yaml v4
// representation nodes. Every function is PURE: inputs are never mutated and results
// are freshly allocated nodes detached from any document. For in-place mutation and
// merge, use the companion package go-yaml-dom.
//
// The first node binds to "." and $1; subsequent nodes bind to $2, $3, ...
package yq

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"

	"github.com/yaml/go-yaml-yq/internal/yqlib"
	yaml "go.yaml.in/yaml/v4"
)

func listOf(c *yqlib.CandidateNode) *list.List {
	l := list.New()
	l.PushBack(c)
	return l
}

// Nodes evaluates a yq expression and returns the full result stream.
func Nodes(expr string, nodes ...*yaml.Node) ([]*yaml.Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("yq: Nodes requires at least one node")
	}
	yqlib.InitExpressionParser()

	cands := make([]*yqlib.CandidateNode, len(nodes))
	for i, n := range nodes {
		if n == nil {
			return nil, fmt.Errorf("yq: node %d is nil", i+1)
		}
		var c yqlib.CandidateNode
		if err := c.UnmarshalYAML(n, map[string]*yqlib.CandidateNode{}); err != nil {
			return nil, fmt.Errorf("yq: converting node %d: %w", i+1, err)
		}
		cands[i] = &c
	}

	exprNode, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err != nil {
		return nil, fmt.Errorf("yq: parsing expression: %w", err)
	}

	ctx := yqlib.Context{MatchingNodes: listOf(cands[0])} // "." = node1
	for i, c := range cands {
		ctx.SetVariable(strconv.Itoa(i+1), listOf(c)) // $1, $2, ...
	}

	out, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(ctx, exprNode)
	if err != nil {
		return nil, fmt.Errorf("yq: evaluating expression: %w", err)
	}

	var results []*yaml.Node
	for e := out.MatchingNodes.Front(); e != nil; e = e.Next() {
		cn, ok := e.Value.(*yqlib.CandidateNode)
		if !ok {
			return nil, errors.New("yq: unexpected result node type")
		}
		yn, err := cn.MarshalYAML()
		if err != nil {
			return nil, fmt.Errorf("yq: converting result: %w", err)
		}
		results = append(results, yn)
	}
	return results, nil
}

// Node errors unless the expression yields exactly one node.
func Node(expr string, nodes ...*yaml.Node) (*yaml.Node, error) {
	rs, err := Nodes(expr, nodes...)
	if err != nil {
		return nil, err
	}
	if len(rs) != 1 {
		return nil, fmt.Errorf("yq: expected exactly one result, got %d", len(rs))
	}
	return rs[0], nil
}
```

**Verify:** `go build ./...`, `go vet ./...` pass.

---

## 5. Licensing / attribution

- yq is MIT-licensed. **Preserve original license headers verbatim** in copied files.
- Add a top-level `NOTICE` stating `internal/yqlib/` contains code copied from
  `github.com/mikefarah/yq` (record exact commit/version) under the MIT License, with
  yq's copyright and license text.

---

## 6. README

- One-line: a Go binding for yq's expression engine over go-yaml v4 nodes.
- Install line.
- **State the purity contract early:** every function returns new nodes, never mutates
  inputs; for in-place editing and merge, use go-yaml-dom.
- Examples: `yq.Node(".a.b", doc)`; `yq.Nodes(".items[]", doc)`; set/delete as
  expressions — `yq.Node(".a.b = $2", doc, v)`, `yq.Node("del(.a.b)", doc)`; the
  binding contract. (Merge is `dom.Merge`, not here.)
- A "works with go-yaml-dom" combined example.
- Link to yq's operator docs; note the engine is vendored yq code (attribution).
- "Status: prototype."

---

## 7. Tests

1. **Identity round-trip.** `Node(".", n)` / `Node("$1", n)` re-emits byte-for-byte
   equal YAML. Corpus: head/line/foot comments, block + flow styles, anchors/aliases,
   explicit `!!merge <<`, multi-line scalars, custom tags.
2. **Purity.** After `Nodes`, re-encode inputs and assert unchanged.
3. **Set/delete via expressions.** `Node(".a.b = $2", doc, v)` → copy with `.a.b == v`,
   `doc` unchanged; `Node("del(.a)", doc)` → copy without `.a`, `doc` unchanged.
4. **Cardinality / errors.** `Nodes(".a, .b", n)` → 2; `Node(".[]", arr)` errors when
   len ≠ 1; malformed expression → parse error, not panic; zero nodes → error.
5. **Engine-merge sanity (optional).** `Nodes("$1 * $2", a, b)` performs a merge via
   the engine, confirming the binding works for arbitrary operators — but the
   *supported* merge API is `dom.Merge`.

**Verify (prototype acceptance):** `go test ./...` passes, with tests 1 and 2 green.

---

## 8. Out of scope / deferred

- **Code-sharing with yq upstream.** After the prototype works, propose exposing the
  engine as an importable core module (format adapters split into separate modules),
  then replace `internal/yqlib/` with that dependency — a one-line swap.
- **Family portability:** this module is Go-specific (it embeds a Go expression
  engine) and is **not** expected to port. The portable surface (including merge)
  lives in go-yaml-dom.

---

## 9. Appendix — verified engine facts

- **Init:** `yqlib.InitExpressionParser()` once before evaluating (lazy; safe to
  repeat).
- **Parse:** `yqlib.ExpressionParser.ParseExpression(expr) (*ExpressionNode, error)`.
- **Evaluate:** `yqlib.NewDataTreeNavigator().GetMatchingNodes(ctx, exprNode)
  (Context, error)`; results in `Context.MatchingNodes` (`*container/list.List` of
  `*CandidateNode`).
- **Inputs:** `Context.MatchingNodes` = one-element list (node1) for `.`;
  `SetVariable("1", …)`, `("2", …)` for `$1`, `$2`, … Each value is a `*list.List`
  with one `*CandidateNode`.
- **Bridge:** `UnmarshalYAML(*yaml.Node, map[string]*CandidateNode)` deep-copies in
  (fresh map per node); `MarshalYAML() (*yaml.Node, error)` emits fresh nodes — this
  boundary is what makes the package pure.
- **Variable names:** lexer rule `\$[a-zA-Z_\-0-9]+`; `$` stripped, `$1` keyed `"1"`.
- **Format build tags:** `yq_nohcl`, `yq_notoml`, `yq_nolua`, `yq_noini`,
  `yq_noprops`, `yq_nocsv`, `yq_noxml`, `yq_nojson`, `yq_nobase64`, `yq_nouri`,
  `yq_nosh`, `yq_noshell`, `yq_nokyaml`. No `yq_nogoccy` — remove goccy by deleting
  `color_print.go` and the goccy decoder/candidate files.
- **Version pin:** engine copied against `go.yaml.in/yaml/v4 v4.0.0-rc.5`; require the
  identical version (and match go-yaml-dom — §1.3).
