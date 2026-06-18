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
