package yq_test

import (
	"fmt"

	yq "github.com/yaml/go-yaml-yq"
	yaml "go.yaml.in/yaml/v4"
)

func mustNode(src string) *yaml.Node {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		panic(err)
	}
	return doc.Content[0]
}

func ExampleNode() {
	doc := mustNode("a:\n  b: value\n")
	got, err := yq.Node(".a.b", doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(got.Value)
	// Output: value
}

func ExampleNodes() {
	doc := mustNode("items: [a, b]\n")
	got, err := yq.Nodes(".items[]", doc)
	if err != nil {
		panic(err)
	}
	for _, node := range got {
		fmt.Println(node.Value)
	}
	// Output:
	// a
	// b
}
