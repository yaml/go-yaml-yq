package yq

import (
	"bytes"
	"fmt"
	"io"

	yaml "go.yaml.in/yaml/v4"
)

type tagDirectiveInfo struct {
	Handle string `yaml:"handle"`
	Prefix string `yaml:"prefix"`
}

type dumpMapItem struct {
	Key   string
	Value any
}

type dumpMapSlice []dumpMapItem

func (ms dumpMapSlice) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, item := range ms {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: item.Key,
		}
		valueNode := &yaml.Node{}
		if err := valueNode.Dump(item.Value); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valueNode)
	}
	return node, nil
}

type nodeInfo struct {
	Kind          string             `yaml:"kind"`
	Style         string             `yaml:"style,omitempty"`
	Anchor        string             `yaml:"anchor,omitempty"`
	Tag           string             `yaml:"tag,omitempty"`
	Head          string             `yaml:"head,omitempty"`
	Line          string             `yaml:"line,omitempty"`
	Foot          string             `yaml:"foot,omitempty"`
	Text          string             `yaml:"text,omitempty"`
	Content       []*nodeInfo        `yaml:"content,omitempty"`
	Encoding      string             `yaml:"encoding,omitempty"`
	Version       string             `yaml:"version,omitempty"`
	TagDirectives []tagDirectiveInfo `yaml:"tag-directives,omitempty"`
}

// DumpNode returns a compact YAML dump of node's go-yaml node structure.
func DumpNode(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteNodeDump(&buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DumpNodeLong returns a profuse YAML dump of node's go-yaml node structure.
func DumpNodeLong(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteNodeDumpLong(&buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteNodeDump writes a compact YAML dump of node's go-yaml node structure.
func WriteNodeDump(w io.Writer, node *yaml.Node) error {
	if w == nil {
		return fmt.Errorf("yq: writer is nil")
	}
	if node == nil {
		return fmt.Errorf("yq: node is nil")
	}
	out, err := yaml.Marshal(formatNodeCompact(*node))
	if err != nil {
		return fmt.Errorf("yq: dumping node: %w", err)
	}
	_, err = w.Write(out)
	return err
}

// WriteNodeDumpLong writes a profuse YAML dump of node's go-yaml node structure.
func WriteNodeDumpLong(w io.Writer, node *yaml.Node) error {
	if w == nil {
		return fmt.Errorf("yq: writer is nil")
	}
	if node == nil {
		return fmt.Errorf("yq: node is nil")
	}
	out, err := yaml.Marshal(formatNode(*node, true))
	if err != nil {
		return fmt.Errorf("yq: dumping node: %w", err)
	}
	_, err = w.Write(out)
	return err
}

func formatNode(n yaml.Node, profuse bool) *nodeInfo {
	info := &nodeInfo{Kind: formatKind(n.Kind)}

	if n.Kind != yaml.DocumentNode && n.Kind != yaml.StreamNode {
		if style := formatStyle(n.Style, profuse); style != "" {
			info.Style = style
		}
	}
	if n.Anchor != "" {
		info.Anchor = n.Anchor
	}
	if tag := formatTag(n.Tag, n.Style, profuse); tag != "" {
		info.Tag = tag
	}
	if n.HeadComment != "" {
		info.Head = n.HeadComment
	}
	if n.LineComment != "" {
		info.Line = n.LineComment
	}
	if n.FootComment != "" {
		info.Foot = n.FootComment
	}

	if info.Kind == "Scalar" {
		info.Text = n.Value
	} else if n.Content != nil {
		info.Content = make([]*nodeInfo, len(n.Content))
		for i, node := range n.Content {
			info.Content[i] = formatNode(*node, profuse)
		}
	}

	if info.Kind == "Stream" && n.Stream != nil {
		if n.Stream.Encoding != 0 {
			info.Encoding = formatEncoding(n.Stream.Encoding)
		}
		if n.Stream.Version != nil {
			info.Version = formatVersion(n.Stream.Version)
		}
		if len(n.Stream.TagDirectives) > 0 {
			info.TagDirectives = make([]tagDirectiveInfo, len(n.Stream.TagDirectives))
			for i, td := range n.Stream.TagDirectives {
				info.TagDirectives[i] = tagDirectiveInfo{
					Handle: td.Handle,
					Prefix: td.Prefix,
				}
			}
		}
	}

	return info
}

func formatNodeCompact(n yaml.Node) any {
	switch n.Kind {
	case yaml.DocumentNode:
		hasProperties := n.Anchor != "" || n.HeadComment != "" || n.LineComment != "" || n.FootComment != ""
		if tag := formatTag(n.Tag, n.Style, false); tag != "" && tag != "!!str" {
			hasProperties = true
		}
		if !hasProperties {
			if len(n.Content) > 0 {
				return formatNodeCompact(*n.Content[0])
			}
			return nil
		}

		result := dumpMapSlice{}
		result = appendNodeProps(result, n, false)
		if len(n.Content) > 0 {
			if contentMap, ok := formatNodeCompact(*n.Content[0]).(dumpMapSlice); ok {
				result = append(result, contentMap...)
			}
		}
		return result
	case yaml.MappingNode:
		result := appendNodeProps(dumpMapSlice{}, n, false)
		var content []any
		for _, node := range n.Content {
			content = append(content, formatNodeCompact(*node))
		}
		return append(result, dumpMapItem{Key: "mapping", Value: content})
	case yaml.SequenceNode:
		result := appendNodeProps(dumpMapSlice{}, n, false)
		var content []any
		for _, node := range n.Content {
			content = append(content, formatNodeCompact(*node))
		}
		return append(result, dumpMapItem{Key: "sequence", Value: content})
	case yaml.ScalarNode:
		result := appendNodeProps(dumpMapSlice{}, n, false)
		return append(result, dumpMapItem{Key: formatStyleName(n.Style), Value: n.Value})
	case yaml.AliasNode:
		return dumpMapSlice{{Key: "alias", Value: n.Value}}
	case yaml.StreamNode:
		content := dumpMapSlice{}
		if n.Stream != nil {
			if n.Stream.Encoding != 0 {
				content = append(content, dumpMapItem{Key: "encoding", Value: formatEncoding(n.Stream.Encoding)})
			}
			if n.Stream.Version != nil {
				content = append(content, dumpMapItem{Key: "version", Value: formatVersion(n.Stream.Version)})
			}
			if len(n.Stream.TagDirectives) > 0 {
				var directives []tagDirectiveInfo
				for _, td := range n.Stream.TagDirectives {
					directives = append(directives, tagDirectiveInfo{Handle: td.Handle, Prefix: td.Prefix})
				}
				content = append(content, dumpMapItem{Key: "tag-directives", Value: directives})
			}
		}
		if len(content) > 0 {
			return dumpMapSlice{{Key: "stream", Value: content}}
		}
		return dumpMapSlice{{Key: "stream", Value: nil}}
	default:
		return nil
	}
}

func appendNodeProps(result dumpMapSlice, n yaml.Node, profuse bool) dumpMapSlice {
	if n.Anchor != "" {
		result = append(result, dumpMapItem{Key: "anchor", Value: n.Anchor})
	}
	if tag := formatTag(n.Tag, n.Style, profuse); tag != "" && tag != "!!str" {
		result = append(result, dumpMapItem{Key: "tag", Value: tag})
	}
	if n.HeadComment != "" {
		result = append(result, dumpMapItem{Key: "head", Value: n.HeadComment})
	}
	if n.LineComment != "" {
		result = append(result, dumpMapItem{Key: "line", Value: n.LineComment})
	}
	if n.FootComment != "" {
		result = append(result, dumpMapItem{Key: "foot", Value: n.FootComment})
	}
	return result
}

func formatEncoding(e yaml.Encoding) string {
	switch e {
	case yaml.EncodingAny:
		return "Any"
	case yaml.EncodingUTF8:
		return "UTF-8"
	case yaml.EncodingUTF16LE:
		return "UTF-16LE"
	case yaml.EncodingUTF16BE:
		return "UTF-16BE"
	default:
		return "Unknown"
	}
}

func formatVersion(v *yaml.VersionDirective) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func formatKind(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "Document"
	case yaml.SequenceNode:
		return "Sequence"
	case yaml.MappingNode:
		return "Mapping"
	case yaml.ScalarNode:
		return "Scalar"
	case yaml.AliasNode:
		return "Alias"
	case yaml.StreamNode:
		return "Stream"
	default:
		return "Unknown"
	}
}

func formatStyle(s yaml.Style, profuse bool) string {
	switch s &^ yaml.TaggedStyle {
	case yaml.DoubleQuotedStyle:
		return "Double"
	case yaml.SingleQuotedStyle:
		return "Single"
	case yaml.LiteralStyle:
		return "Literal"
	case yaml.FoldedStyle:
		return "Folded"
	case yaml.FlowStyle:
		return "Flow"
	case 0:
		if profuse {
			return "Plain"
		}
	}
	return ""
}

func formatStyleName(s yaml.Style) string {
	switch s &^ yaml.TaggedStyle {
	case yaml.DoubleQuotedStyle:
		return "double"
	case yaml.SingleQuotedStyle:
		return "single"
	case yaml.LiteralStyle:
		return "literal"
	case yaml.FoldedStyle:
		return "folded"
	case yaml.FlowStyle:
		return "flow"
	case 0:
		return "plain"
	default:
		return "unknown"
	}
}

func formatTag(tag string, style yaml.Style, profuse bool) string {
	tagWasExplicit := style&yaml.TaggedStyle != 0
	if profuse {
		return tag
	}
	switch tag {
	case "!!str", "!!map", "!!seq", "!!int", "!!float", "!!bool", "!!null":
		if tagWasExplicit {
			return tag
		}
		return ""
	}
	return tag
}
