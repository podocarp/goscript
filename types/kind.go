package types

import (
	"reflect"

	"github.com/go-errors/errors"
)

type Type interface {
	Kind() Kind

	// Elem returns a type's element type.
	// The type's Kind must be Array.
	Elem() (Type, error)
	String() string
	Equal(Type) bool
}

type Kind uint

const (
	Invalid Kind = iota
	String
	Float
	Int

	Array
	Func
	Builtin
)

var kindStr = []string{
	Invalid: "invalid",
	String:  "string",
	Float:   "float",
	Int:     "int",
	Array:   "array",
	Func:    "function",
	Builtin: "builtin",
}

func (k Kind) String() string {
	return kindStr[k]
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

// Predefined types for convenience
var (
	StringType Type = LiteralOf(String)
	FloatType       = LiteralOf(Float)
	IntType         = LiteralOf(Int)
	// TODO: func should contain parameter's types
	FuncType    = LiteralOf(Func)
	BuiltinType = LiteralOf(Builtin)
)

type _type struct {
	kind Kind
	elt  Type
}

func LiteralOf(kind Kind) *_type {
	return &_type{
		kind: kind,
	}
}

func ArrayOf(elemType Type) *_type {
	return &_type{
		kind: Array,
		elt:  elemType,
	}
}

func (t *_type) Kind() Kind {
	return t.kind
}

func (t *_type) Elem() (Type, error) {
	if t.kind != Array {
		return nil, errors.New("cannot call Elem for non-array type")
	}

	return t.elt, nil
}

func (t *_type) String() string {
	switch t.kind {
	case Array:
		return "[]" + t.elt.String()
	default:
		return t.kind.String()
	}
}

func (t *_type) Equal(other Type) bool {
	if t.Kind() != other.Kind() {
		return false
	}

	if t.Kind() == Array {
		otherElem, _ := other.Elem()
		return t.elt.Equal(otherElem)
	} else {
		return true
	}
}

func ReflectTypeToType(r reflect.Type) (*_type, error) {
	if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
		// new literal of the same type
		kind, err := ReflectKindToKind(r.Kind())
		if err != nil {
			return nil, err
		}
		return LiteralOf(kind), nil
	}

	if r.Elem().Kind() == reflect.Array || r.Elem().Kind() == reflect.Slice {
		// array of array
		elemType, err := ReflectTypeToType(r.Elem())
		if err != nil {
			return nil, err
		}
		return ArrayOf(elemType), nil
	} else {
		// new array of the same type
		kind, err := ReflectKindToKind(r.Elem().Kind())
		if err != nil {
			return nil, err
		}
		return ArrayOf(LiteralOf(kind)), nil
	}
}
