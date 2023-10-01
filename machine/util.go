package machine

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/podocarp/goscript/kind"
)

func isTruthyFloat(val float64) bool {
	if val > 0 {
		return true
	}
	return false
}

func isTruthy(node *Node) bool {
	switch node.Kind {
	case kind.FLOAT:
		val, _ := node.Value.(float64)
		return isTruthyFloat(val)
	case kind.STRING:
		if node.Value.(string) == "" {
			return false
		}
		return true
	case kind.FUNC:
		return true
	}
	return false
}

func (m *machine) AstNodeToNode(lit ast.Node) *Node {
	var val any
	switch n := lit.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.FLOAT, token.INT:
			val, _ = strconv.ParseFloat(n.Value, 64)
			return &Node{
				Kind:  kind.FLOAT,
				Value: val,
			}
		case token.CHAR, token.STRING:
			return &Node{
				Kind:  kind.STRING,
				Value: n.Value,
			}
		}
	case *ast.FuncLit:
		return &Node{
			Kind:    kind.FUNC,
			Value:   val,
			Context: m.Context,
		}
	}

	return nil
}
