// Package yq provides yq expression evaluation over go-yaml v4 representation
// nodes.
//
// The package works directly with [yaml.Node] values from go.yaml.in/yaml/v4.
// Node and Nodes are pure: they never mutate input nodes, and every returned
// node is a detached copy rather than an interior pointer into the source
// document.
//
// Evaluate a query:
//
//	var doc yaml.Node
//	_ = yaml.Unmarshal([]byte("name: demo\n"), &doc)
//	name, err := yq.Node(".name", doc.Content[0])
//
// Evaluate an update without mutating the original:
//
//	updated, err := yq.Node(".image = \"example/app:2.0\"", root)
//
// The first input node binds to both "." and $1. Additional input nodes bind to
// $2, $3, and so on:
//
//	merged, err := yq.Node("$1 * $2", base, overlay)
//
// Render helpers print node graphs with yq's YAML and JSON encoders:
//
//	err := yq.WriteYAML(os.Stdout, updated, yq.WithColor(true))
//	err := yq.WriteJSON(os.Stdout, updated, yq.WithColor(true))
//
// Use github.com/yaml/go-yaml-dom when you want live structural mutation. The
// two packages compose only through *yaml.Node; neither imports the other.
package yq

import yaml "go.yaml.in/yaml/v4"

var _ *yaml.Node
