package yq

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"

	"github.com/yaml/go-yaml-yq/internal/yqlib"
	yaml "go.yaml.in/yaml/v4"
)

func listOf(c *yqlib.CandidateNode) *list.List {
	l := list.New()
	l.PushBack(c)
	return l
}

func candidateFromNode(n *yaml.Node, index int) (*yqlib.CandidateNode, error) {
	if n == nil {
		if index > 0 {
			return nil, fmt.Errorf("yq: node %d is nil", index)
		}
		return nil, errors.New("yq: node is nil")
	}
	var c yqlib.CandidateNode
	if err := c.UnmarshalYAML(n, map[string]*yqlib.CandidateNode{}); err != nil {
		if index > 0 {
			return nil, fmt.Errorf("yq: converting node %d: %w", index, err)
		}
		return nil, fmt.Errorf("yq: converting node: %w", err)
	}
	return &c, nil
}

// Nodes evaluates a yq expression and returns the full result stream.
func Nodes(expr string, nodes ...*yaml.Node) ([]*yaml.Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("yq: Nodes requires at least one node")
	}
	yqlib.InitExpressionParser()

	cands := make([]*yqlib.CandidateNode, len(nodes))
	for i, n := range nodes {
		c, err := candidateFromNode(n, i+1)
		if err != nil {
			return nil, err
		}
		cands[i] = c
	}

	exprNode, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err != nil {
		return nil, fmt.Errorf("yq: parsing expression: %w", err)
	}

	ctx := yqlib.Context{MatchingNodes: listOf(cands[0])}
	for i, c := range cands {
		ctx.SetVariable(strconv.Itoa(i+1), listOf(c))
	}

	out, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(ctx, exprNode)
	if err != nil {
		return nil, fmt.Errorf("yq: evaluating expression: %w", err)
	}

	var results []*yaml.Node
	for e := out.MatchingNodes.Front(); e != nil; e = e.Next() {
		cn, ok := e.Value.(*yqlib.CandidateNode)
		if !ok {
			return nil, errors.New("yq: unexpected result node type")
		}
		yn, err := cn.MarshalYAML()
		if err != nil {
			return nil, fmt.Errorf("yq: converting result: %w", err)
		}
		results = append(results, yn)
	}
	return results, nil
}

// Node evaluates a yq expression and errors unless it yields exactly one node.
func Node(expr string, nodes ...*yaml.Node) (*yaml.Node, error) {
	rs, err := Nodes(expr, nodes...)
	if err != nil {
		return nil, err
	}
	if len(rs) != 1 {
		return nil, fmt.Errorf("yq: expected exactly one result, got %d", len(rs))
	}
	return rs[0], nil
}
