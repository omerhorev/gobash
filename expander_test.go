package gobash

import (
	"testing"

	"github.com/omerhorev/gobash/ast"
	"github.com/stretchr/testify/require"
)

func TestExpanderBacktickRecursive(t *testing.T) {
	testExpander(t, "x`y`", ast.NewExpr(
		ast.String{Value: "x"},
		ast.Backtick{Node: ast.NewExpr(
			ast.String{Value: "y"},
		)},
	))
}

func testExpander(t *testing.T, expr string, node ast.Node) {
	e := NewExpander(expr)
	require.NoError(t, e.Parse())

	nodesEqual(node, e.Expr)
}

func nodesEqual(n1 ast.Node, n2 ast.Node) bool {
	switch node := n1.(type) {
	case *ast.String:
		if node2, ok := n2.(ast.String); !ok {
			return false
		} else {
			return node.Value == node2.Value
		}
	case *ast.Expr:
		if node2, ok := n2.(ast.Expr); !ok {
			return false
		} else {
			if len(node.Nodes) != len(node2.Nodes) {
				return false
			}

			for i := range node.Nodes {
				if !nodesEqual(node.Nodes[i], node2.Nodes[i]) {
					return false
				}
			}

			return true
		}
	case *ast.Backtick:
		if node2, ok := n2.(ast.Backtick); !ok {
			return false
		} else {
			return nodesEqual(node.Node, node2.Node)
		}
	}

	return false
}
