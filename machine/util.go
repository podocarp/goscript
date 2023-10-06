package machine

import (
	"github.com/podocarp/goscript/types"
)

func isTruthyFloat(val float64) bool {
	if val > 0 {
		return true
	}
	return false
}

func isTruthy(node *Node) bool {
	switch node.Type {
	case types.FloatType:
		val := node.Value.(Number).ToFloat()
		return isTruthyFloat(val)
	case types.StringType:
		if node.Value.(string) == "" {
			return false
		}
		return true
	case types.FuncType:
		return true
	}
	return false
}
