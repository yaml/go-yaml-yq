package yq

import (
	"bytes"
	"strings"
	"testing"
)

func TestDumpNodeCompact(t *testing.T) {
	node := parseRoot(t, "name: demo\nitems:\n  - a\n  - b\n")

	out, err := DumpNode(node)
	if err != nil {
		t.Fatalf("DumpNode: %v", err)
	}
	got := string(out)
	for _, want := range []string{"mapping:", "plain: name", "plain: demo", "plain: items", "sequence:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("DumpNode missing %q in:\n%s", want, got)
		}
	}
}

func TestDumpNodeLong(t *testing.T) {
	node := parseRoot(t, "name: demo\n")

	out, err := DumpNodeLong(node)
	if err != nil {
		t.Fatalf("DumpNodeLong: %v", err)
	}
	got := string(out)
	for _, want := range []string{"kind: Mapping", "style: Plain", "tag: '!!map'", "kind: Scalar", "text: name"} {
		if !strings.Contains(got, want) {
			t.Fatalf("DumpNodeLong missing %q in:\n%s", want, got)
		}
	}
}

func TestDumpNodeCommentsAndStyles(t *testing.T) {
	node := parseRoot(t, "# head\nname: \"demo\" # line\n")

	out, err := DumpNode(node)
	if err != nil {
		t.Fatalf("DumpNode: %v", err)
	}
	got := string(out)
	for _, want := range []string{"head: '# head'", "line: '# line'", "double: demo"} {
		if !strings.Contains(got, want) {
			t.Fatalf("DumpNode missing %q in:\n%s", want, got)
		}
	}
}

func TestWriteNodeDumpErrors(t *testing.T) {
	if _, err := DumpNode(nil); err == nil {
		t.Fatalf("DumpNode nil should error")
	}
	if err := WriteNodeDump(nil, parseRoot(t, "a: b\n")); err == nil {
		t.Fatalf("WriteNodeDump nil writer should error")
	}

	var buf bytes.Buffer
	if err := WriteNodeDumpLong(&buf, parseRoot(t, "a: b\n")); err != nil {
		t.Fatalf("WriteNodeDumpLong: %v", err)
	}
	if !strings.Contains(buf.String(), "kind: Mapping") {
		t.Fatalf("unexpected WriteNodeDumpLong output:\n%s", buf.String())
	}
}
