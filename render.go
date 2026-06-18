package yq

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yaml/go-yaml-yq/internal/yqlib"
	yaml "go.yaml.in/yaml/v4"
)

type renderOptions struct {
	color        bool
	indent       int
	unwrapScalar bool
}

// RenderOption configures YAML and JSON rendering.
type RenderOption func(*renderOptions)

// WithColor controls ANSI color in rendered output. Color is enabled by default.
func WithColor(enabled bool) RenderOption {
	return func(o *renderOptions) {
		o.color = enabled
	}
}

// WithIndent controls YAML and JSON indentation. It must be between 2 and 9.
func WithIndent(indent int) RenderOption {
	return func(o *renderOptions) {
		o.indent = indent
	}
}

// WithUnwrapScalar controls yq's scalar unwrapping behavior.
func WithUnwrapScalar(enabled bool) RenderOption {
	return func(o *renderOptions) {
		o.unwrapScalar = enabled
	}
}

func defaultRenderOptions() renderOptions {
	return renderOptions{
		color:        true,
		indent:       2,
		unwrapScalar: true,
	}
}

func applyRenderOptions(opts []RenderOption) (renderOptions, error) {
	cfg := defaultRenderOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.indent < 2 || cfg.indent > 9 {
		return cfg, fmt.Errorf("yq: indent must be between 2 and 9, got %d", cfg.indent)
	}
	return cfg, nil
}

// RenderYAML renders node as YAML using yq's YAML encoder.
func RenderYAML(node *yaml.Node, opts ...RenderOption) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteYAML(&buf, node, opts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteYAML writes node as YAML using yq's YAML encoder.
func WriteYAML(w io.Writer, node *yaml.Node, opts ...RenderOption) error {
	if w == nil {
		return errorsNewNilWriter()
	}
	cfg, err := applyRenderOptions(opts)
	if err != nil {
		return err
	}
	c, err := candidateFromNode(node, 0)
	if err != nil {
		return err
	}
	prefs := yqlib.ConfiguredYamlPreferences.Copy()
	prefs.ColorsEnabled = cfg.color
	prefs.Indent = cfg.indent
	prefs.UnwrapScalar = cfg.unwrapScalar
	return yqlib.NewYamlEncoder(prefs).Encode(w, c)
}

// RenderJSON renders node as JSON using yq's JSON encoder.
func RenderJSON(node *yaml.Node, opts ...RenderOption) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, node, opts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteJSON writes node as JSON using yq's JSON encoder.
func WriteJSON(w io.Writer, node *yaml.Node, opts ...RenderOption) error {
	if w == nil {
		return errorsNewNilWriter()
	}
	cfg, err := applyRenderOptions(opts)
	if err != nil {
		return err
	}
	c, err := candidateFromNode(node, 0)
	if err != nil {
		return err
	}
	prefs := yqlib.ConfiguredJSONPreferences.Copy()
	prefs.ColorsEnabled = cfg.color
	prefs.Indent = cfg.indent
	prefs.UnwrapScalar = cfg.unwrapScalar
	return yqlib.NewJSONEncoder(prefs).Encode(w, c)
}

func errorsNewNilWriter() error {
	return fmt.Errorf("yq: writer is nil")
}
