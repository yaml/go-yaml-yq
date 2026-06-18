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
