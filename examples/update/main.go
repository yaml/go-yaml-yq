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
