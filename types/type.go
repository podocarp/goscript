package types

import (
	"reflect"

	"github.com/go-errors/errors"
)

// Predefined types for convenience
var (
	StringType Type = LiteralOf(String)
	FloatType       = LiteralOf(Float)
	IntType         = LiteralOf(Int)
	UintType        = LiteralOf(Uint)
	BoolType        = LiteralOf(Bool)
	// TODO: func should contain parameter's types
	FuncType    = LiteralOf(Func)
	BuiltinType = LiteralOf(Builtin)

	// The so called idiom on go/reflect's pkg.go.dev page:
	// reflect.TypeOf((*string)(nil)).Elem()
	// is actually 2x slower than just initializing the object for small
	// literals
	stringReflectType = reflect.TypeOf("")
	floatReflectType  = reflect.TypeOf(float64(0))
	intReflectType    = reflect.TypeOf(int64(0))
	uintReflectType   = reflect.TypeOf(uint64(0))
)

type Type interface {
	// Kind returns the underlying Kind of this type
	Kind() Kind

	// Elem returns a type's element type.
	// The type's Kind must be Array.
	Elem() (Type, error)
	Equal(Type) bool

	String() string
	TypeToReflectType() reflect.Type
}

type _type struct {
	kind  Kind
	elt   Type
	islit bool
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

func (t *_type) TypeToReflectType() reflect.Type {
	switch t.kind {
	case Array:
		elementType := t.elt.TypeToReflectType()
		return reflect.SliceOf(elementType)
	case Float:
		return floatReflectType
	case Func:
		return reflect.FuncOf([]reflect.Type{}, []reflect.Type{}, false)
	case Int:
		return intReflectType
	case String:
		return stringReflectType
	case Uint:
		return intReflectType
	default:
		return nil
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
