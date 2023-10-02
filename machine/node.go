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
	switch n.Kind {
	case kind.ARRAY:
		return arrToString(n.Value.([]*Node))
	case kind.FLOAT:
		return fmt.Sprint(n.Value)
	case kind.STRING:
		return strconv.Quote(fmt.Sprint(n.Value))
	case kind.FUNC:
		return "λ"
	default:
		return "unknown type"
	}
}
