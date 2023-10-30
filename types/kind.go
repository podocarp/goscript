package types

import (
	"reflect"

	"github.com/go-errors/errors"
)

type Kind uint

const (
	Invalid Kind = iota
	String
	Float
	Int
	Uint
	Bool

	Array
	Func
	Builtin
	// Used for multiple value statements, like a, b := f()
	Packing
)

var kindStr = []string{
	Invalid: "invalid",
	String:  "string",
	Float:   "float",
	Int:     "int",
	Uint:    "uint",
	Bool:    "bool",

	Array:   "array",
	Func:    "function",
	Builtin: "builtin",
	Packing: "packing",
}

func (k Kind) String() string {
	return kindStr[k]
}

func (k Kind) IsNumeric() bool {
	return k == Float || k == Int || k == Uint
}

func ReflectKindToKind(r reflect.Kind) (Kind, error) {
	switch r {
	case reflect.String:
		return String, nil
	case reflect.Float32, reflect.Float64:
		return Float, nil
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return Int, nil
	case reflect.Array, reflect.Slice:
		return Array, nil
	default:
		return Invalid, errors.Errorf("unsupported reflect.Kind %s", r)
	}
}
