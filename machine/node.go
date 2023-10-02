package machine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/podocarp/goscript/kind"
)

type Node struct {
	Kind    kind.Kind
	Value   any
	Context *context

	IsReturnValue bool
}

func arrToString(arr []*Node) string {
	var arrContents strings.Builder
	arrContents.WriteString("[ ")
	for _, elem := range arr {
		if elem.Kind == kind.ARRAY {
			arrContents.WriteString(arrToString(elem.Value.([]*Node)))
		} else {
			arrContents.WriteString(strconv.Quote(fmt.Sprint(elem.Value)))
		}
		arrContents.WriteString(" ")
	}
	arrContents.WriteString("]")

	return arrContents.String()
}

func (n *Node) String() string {
	var val string
	switch n.Kind {
	case kind.ARRAY:
		val = arrToString(n.Value.([]*Node))
	case kind.FLOAT:
		val = fmt.Sprint(n.Value)
	case kind.STRING:
		val = strconv.Quote(fmt.Sprint(n.Value))
	case kind.FUNC:
		val = "λ"
	default:
		return "unknown type"
	}

	if n.IsReturnValue {
		return val + "[r]"
	}
	return val
}
