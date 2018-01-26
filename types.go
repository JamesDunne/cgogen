// types.go
package main

import (
	"fmt"
	"strings"

	"github.com/cznic/cc"
	"github.com/cznic/xc"
)

type Type struct {
	cc.Type
}

func Export(name string) string {
	return strings.Title(name)
}

func goName(name string) string {
	if strings.HasPrefix(name, "VG") {
		return Export(name[2:])
	}
	return Export(name)
}

func (t Type) IsConst() bool {
	return t.Specifier().IsConst()
}

func (t Type) RequiresCast() bool {
	switch t.Kind() {
	case cc.Undefined:
		return false
	case cc.Void:
		return false
	case cc.Ptr:
		return Type{t.Element()}.RequiresCast()
	case cc.UintPtr: // Type used for pointer arithmetic.
		return false
	case cc.Char:
		return false
	case cc.SChar:
		return false
	case cc.UChar:
		return false
	case cc.Short:
		return false
	case cc.UShort:
		return false
	case cc.Int:
		return false
	case cc.UInt:
		return false
	case cc.Long:
		return false
	case cc.ULong:
		return false
	case cc.LongLong:
		return false
	case cc.ULongLong:
		return false
	case cc.Float:
		return false
	case cc.Double:
		return false
	case cc.LongDouble:
		return false
	case cc.Bool:
		return false
	case cc.FloatComplex:
		return false
	case cc.DoubleComplex:
		return false
	case cc.LongDoubleComplex:
		return false
	case cc.Struct:
		return true
	case cc.Union:
		return true
	case cc.Enum:
		return true
	case cc.TypedefName:
		return true
	case cc.Function:
		return true
	case cc.Array:
		return Type{t.Element()}.RequiresCast()
	default:
		return false
	}
}

func (t Type) GoType() string {
	switch t.Kind() {
	case cc.Undefined:
		return "undefined"
	case cc.Void:
		return "void"
	case cc.Ptr:
		return "*" + Type{t.Element()}.GoType()
	case cc.UintPtr: // Type used for pointer arithmetic.
		return "uintptr"
	case cc.Char:
		return "int8"
	case cc.SChar:
		return "int8"
	case cc.UChar:
		return "uint8"
	case cc.Short:
		return "int16"
	case cc.UShort:
		return "uint16"
	case cc.Int:
		return "int32"
	case cc.UInt:
		return "uint32"
	case cc.Long:
		return "int32"
	case cc.ULong:
		return "uint32"
	case cc.LongLong:
		return "int64"
	case cc.ULongLong:
		return "uint64"
	case cc.Float:
		return "float32"
	case cc.Double:
		return "float64"
	case cc.LongDouble:
		return "float64"
	case cc.Bool:
		return "bool"
	case cc.FloatComplex:
		return "complex64"
	case cc.DoubleComplex:
		return "complex128"
	case cc.LongDoubleComplex:
		return "complex128"
	case cc.Struct:
		return "struct"
	case cc.Union:
		return "union"
	case cc.Enum:
		return goName(typedefNameOf(t))
	case cc.TypedefName:
		return goName(typedefNameOf(t))
	case cc.Function:
		return "func"
	case cc.Array:
		return fmt.Sprintf("[%d]%s", t.Elements(), Type{t.Element()}.GoType())
	default:
		return "???"
	}
}

func (t Type) CGoType() string {
	switch t.Kind() {
	case cc.Undefined:
		return "undefined"
	case cc.Void:
		return "void"
	case cc.Ptr:
		return "*" + Type{t.Element()}.CGoType()
	case cc.UintPtr: // Type used for pointer arithmetic.
		return "uintptr"
	case cc.Char:
		return "int8"
	case cc.SChar:
		return "int8"
	case cc.UChar:
		return "uint8"
	case cc.Short:
		return "int16"
	case cc.UShort:
		return "uint16"
	case cc.Int:
		return "int32"
	case cc.UInt:
		return "uint32"
	case cc.Long:
		return "int32"
	case cc.ULong:
		return "uint32"
	case cc.LongLong:
		return "int64"
	case cc.ULongLong:
		return "uint64"
	case cc.Float:
		return "float32"
	case cc.Double:
		return "float64"
	case cc.LongDouble:
		return "float64"
	case cc.Bool:
		return "bool"
	case cc.FloatComplex:
		return "complex64"
	case cc.DoubleComplex:
		return "complex128"
	case cc.LongDoubleComplex:
		return "complex128"
	case cc.Struct:
		return "struct"
	case cc.Union:
		return "union"
	case cc.Enum:
		return fmt.Sprintf("C.%s", typedefNameOf(t))
	case cc.TypedefName:
		return fmt.Sprintf("C.%s", typedefNameOf(t))
	case cc.Function:
		return "func"
	case cc.Array:
		return fmt.Sprintf("[%d]%s", t.Elements(), Type{t.Element()}.CGoType())
	default:
		return "???"
	}
}

type Parameter struct {
	identifier string
	Type       Type
}

func (p Parameter) CName() string  { return p.identifier }
func (p Parameter) GoName() string { return p.identifier }

type Function struct {
	identifier string
	Parameters []Parameter
	ResultType Type
}

func (f Function) CName() string { return f.identifier }
func (f Function) GoName() string {
	goName := f.identifier
	if strings.HasPrefix(goName, "vg") {
		goName = goName[2:]
	}
	return strings.Title(goName)
}

func identifierOf(dd *cc.DirectDeclarator) string {
	switch dd.Case {
	case 0: // IDENTIFIER
		if dd.Token.Val == 0 {
			return ""
		}
		return string(dd.Token.S())
	case 1: // '(' Declarator ')'
		return identifierOf(dd.Declarator.DirectDeclarator)
	default:
		//	DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
		//	DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'
		//	DirectDeclarator '[' TypeQualifierList "static" Expression ']'
		//	DirectDeclarator '[' TypeQualifierListOpt '*' ']'
		//	DirectDeclarator '(' ParameterTypeList ')'
		//	DirectDeclarator '(' IdentifierListOpt ')'
		return identifierOf(dd.DirectDeclarator)
	}
}

func typedefNameOf(typ cc.Type) string {
	rawSpec := typ.Declarator().RawSpecifier()
	if name := rawSpec.TypedefName(); name > 0 {
		return string(xc.Dict.S(name))
	} else if rawSpec.IsTypedef() {
		return identifierOf(typ.Declarator().DirectDeclarator)
	}
	return ""
}
