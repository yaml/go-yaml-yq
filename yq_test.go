package yq

import (
	"bytes"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v4"
)

func parseRoot(t *testing.T, src string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected one document root, got %d", len(doc.Content))
	}
	return doc.Content[0]
}

func encode(t *testing.T, n *yaml.Node) []byte {
	t.Helper()
	out, err := yaml.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return out
}

func TestNodeIdentityAndVariableRoundTrip(t *testing.T) {
	src := "# head\na: &base\n  b: 1 # line\n  c: |\n    hello\n    world\nref: *base\n<<: *base\n"
	node := parseRoot(t, src)

	want := encode(t, node)
	for _, expr := range []string{".", "$1"} {
		got, err := Node(expr, node)
		if err != nil {
			t.Fatalf("Node(%q): %v", expr, err)
		}
		if !bytes.Equal(encode(t, got), want) {
			t.Fatalf("Node(%q) did not round-trip\nwant:\n%s\ngot:\n%s", expr, want, encode(t, got))
		}
	}
}

func TestPurity(t *testing.T) {
	node := parseRoot(t, "a:\n  b: old\nitems: [one, two]\n")
	before := encode(t, node)

	if _, err := Nodes(".a.b = \"new\" | del(.items[0])", node); err != nil {
		t.Fatalf("Nodes: %v", err)
	}
	if !bytes.Equal(encode(t, node), before) {
		t.Fatalf("input mutated\nbefore:\n%s\nafter:\n%s", before, encode(t, node))
	}
}

func TestSetAndDeleteExpressions(t *testing.T) {
	doc := parseRoot(t, "a:\n  b: old\nkeep: true\n")
	value := parseRoot(t, "new\n")

	updated, err := Node(".a.b = $2", doc, value)
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := Node(".a.b", updated)
	if err != nil {
		t.Fatalf("read updated: %v", err)
	}
	if got.Value != "new" {
		t.Fatalf("set value = %q, want new", got.Value)
	}

	deleted, err := Node("del(.a)", doc)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	out := string(encode(t, deleted))
	if strings.Contains(out, "a:") || !strings.Contains(out, "keep: true") {
		t.Fatalf("unexpected delete output:\n%s", out)
	}
	if strings.Contains(string(encode(t, doc)), "new") {
		t.Fatalf("source document was mutated")
	}
}

func TestCardinalityAndErrors(t *testing.T) {
	doc := parseRoot(t, "a: 1\nb: 2\narr: [x, y]\n")

	got, err := Nodes(".a, .b", doc)
	if err != nil {
		t.Fatalf("Nodes cardinality: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}

	if _, err := Node(".arr[]", doc); err == nil {
		t.Fatalf("Node should reject multiple results")
	}
	if _, err := Nodes(".[", doc); err == nil {
		t.Fatalf("malformed expression should error")
	}
	if _, err := Nodes("."); err == nil {
		t.Fatalf("zero nodes should error")
	}
	if _, err := Nodes(".", nil); err == nil {
		t.Fatalf("nil node should error")
	}
}
