package machine

import (
	"fmt"

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
		val := node.Value.(Number).ToFloat()
		fmt.Println(node.Value)
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
