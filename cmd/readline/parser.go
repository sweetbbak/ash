package main

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func visit(node syntax.Node) bool {
	switch node := node.(type) {
	case *syntax.Assign:
	case *syntax.ParamExp:
		if node.Slice == nil {
			break
		}
	case *syntax.ArithmExp:
	case *syntax.ArithmCmd:
	case *syntax.ParenArithm:
	case *syntax.BinaryArithm:
	case *syntax.CmdSubst:
	case *syntax.Subshell:
	case *syntax.Word:
	case *syntax.TestClause:
	case *syntax.ParenTest:
	case *syntax.BinaryTest:
	case *syntax.UnaryTest:
	default:
	}

	return true
}

func (a *ash) Highlight(line []rune) string {
	var node syntax.Node
	node, _ = a.parser.Parse(strings.NewReader(string(line)), "highlite")

	syntax.Walk(node, func(n syntax.Node) bool {
		return true
	})

	return string(line)
}
