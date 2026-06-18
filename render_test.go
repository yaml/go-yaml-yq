package yq

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func hasANSI(s string) bool {
	return ansiRE.MatchString(s)
}

func TestRenderYAML(t *testing.T) {
	node := parseRoot(t, "name: demo\nreplicas: 2\nflag: true\n")

	out, err := RenderYAML(node, WithColor(false))
	if err != nil {
		t.Fatalf("RenderYAML: %v", err)
	}
	got := string(out)
	if hasANSI(got) {
		t.Fatalf("RenderYAML WithColor(false) emitted ANSI: %q", got)
	}
	for _, want := range []string{"name: demo", "replicas: 2", "flag: true"} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderYAML missing %q in:\n%s", want, got)
		}
	}

	colored, err := RenderYAML(node)
	if err != nil {
		t.Fatalf("RenderYAML colored: %v", err)
	}
	if !hasANSI(string(colored)) {
		t.Fatalf("RenderYAML default should emit ANSI color")
	}
	if stripANSI(string(colored)) != got {
		t.Fatalf("colored output changed content\nplain:\n%s\ncolored stripped:\n%s", got, stripANSI(string(colored)))
	}
}

func TestRenderJSON(t *testing.T) {
	node := parseRoot(t, "name: demo\nitems: [a, b]\n")

	out, err := RenderJSON(node, WithColor(false))
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	got := string(out)
	if hasANSI(got) {
		t.Fatalf("RenderJSON WithColor(false) emitted ANSI: %q", got)
	}
	for _, want := range []string{`"name": "demo"`, `"items": [`} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderJSON missing %q in:\n%s", want, got)
		}
	}

	colored, err := RenderJSON(node)
	if err != nil {
		t.Fatalf("RenderJSON colored: %v", err)
	}
	if !hasANSI(string(colored)) {
		t.Fatalf("RenderJSON default should emit ANSI color")
	}
	if stripANSI(string(colored)) != got {
		t.Fatalf("colored JSON changed content\nplain:\n%s\ncolored stripped:\n%s", got, stripANSI(string(colored)))
	}
}

func TestRenderUnwrappedScalarColor(t *testing.T) {
	node := parseRoot(t, "true\n")
	out, err := RenderYAML(node)
	if err != nil {
		t.Fatalf("RenderYAML scalar: %v", err)
	}
	if !hasANSI(string(out)) {
		t.Fatalf("unwrapped scalar should be colorized")
	}
	if stripANSI(string(out)) != "true\n" {
		t.Fatalf("scalar content changed: %q", stripANSI(string(out)))
	}
}

func TestRenderOptionsAndErrors(t *testing.T) {
	node := parseRoot(t, "a:\n  b: c\n")

	out, err := RenderJSON(node, WithColor(false), WithIndent(4))
	if err != nil {
		t.Fatalf("RenderJSON indent: %v", err)
	}
	if !strings.Contains(string(out), "\n        \"b\"") {
		t.Fatalf("RenderJSON did not use indent 4:\n%s", out)
	}

	if _, err := RenderYAML(nil); err == nil {
		t.Fatalf("RenderYAML nil should error")
	}
	if _, err := RenderYAML(node, WithIndent(1)); err == nil {
		t.Fatalf("RenderYAML invalid indent should error")
	}
	if err := WriteYAML(nil, node); err == nil {
		t.Fatalf("WriteYAML nil writer should error")
	}
}

func TestWriteRender(t *testing.T) {
	node := parseRoot(t, "a: b\n")
	var buf bytes.Buffer
	if err := WriteYAML(&buf, node, WithColor(false)); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "a: b") {
		t.Fatalf("unexpected WriteYAML output: %q", got)
	}
}
