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

func (t Type) IsTypeDef() bool {
	rawSpec := t.Declarator().RawSpecifier()
	if name := rawSpec.TypedefName(); name > 0 {
		return true
	} else if rawSpec.IsTypedef() {
		return true
	}
	return false
}

func (t Type) RequiresCast() bool {
	if t.IsTypeDef() {
		return true
	}

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
		return "byte"
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
		return goName(typedefNameOf(t)) + "Enum"
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
	prefix := ""
	base := ""

	if t.Kind() == cc.Array {
		prefix = fmt.Sprintf("[%d]", t.Elements())
		t = Type{t.Element()}
	} else if t.Kind() == cc.Ptr {
		prefix = "*"
		t = Type{t.Element()}
	}

	rawSpec := t.Declarator().RawSpecifier()
	if name := rawSpec.TypedefName(); name > 0 {
		base = "C." + blessName(xc.Dict.S(name))
	} else if rawSpec.IsTypedef() {
		base = "C." + identifierOf(t.Declarator().DirectDeclarator)
	} else {
		switch t.Kind() {
		case cc.Undefined:
			base = "undefined"
		case cc.Void:
			base = "byte"
		case cc.UintPtr: // Type used for pointer arithmetic.
			base = "uintptr"
		case cc.Char:
			base = "int8"
		case cc.SChar:
			base = "int8"
		case cc.UChar:
			base = "uint8"
		case cc.Short:
			base = "int16"
		case cc.UShort:
			base = "uint16"
		case cc.Int:
			base = "int32"
		case cc.UInt:
			base = "uint32"
		case cc.Long:
			base = "int32"
		case cc.ULong:
			base = "uint32"
		case cc.LongLong:
			base = "int64"
		case cc.ULongLong:
			base = "uint64"
		case cc.Float:
			base = "float32"
		case cc.Double:
			base = "float64"
		case cc.LongDouble:
			base = "float64"
		case cc.Bool:
			base = "bool"
		case cc.FloatComplex:
			base = "complex64"
		case cc.DoubleComplex:
			base = "complex128"
		case cc.LongDoubleComplex:
			base = "complex128"
		case cc.Struct:
			base = "struct"
		case cc.Union:
			base = "union"
		case cc.Enum:
			base = fmt.Sprintf("C.%s", typedefNameOf(t))
		case cc.TypedefName:
			base = fmt.Sprintf("C.%s", typedefNameOf(t))
		case cc.Function:
			base = "func"
		default:
			base = "???"
		}
	}

	return prefix + base
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

func parseFunction(fnDecl *cc.Declarator) Function {
	fn := fnDecl.DirectDeclarator

	f := Function{
		identifier: identifierOf(fn),
		ResultType: Type{fnDecl.Type.Result()},
	}

	pList := fn.ParameterTypeList.ParameterList
	if pList.ParameterList == nil && pList.ParameterDeclaration.Declarator == nil {
		// empty void parameter list
		// TODO: check for 'void' type?
	} else {
		for pList != nil {
			p := pList.ParameterDeclaration
			if p.Declarator != nil {
				f.Parameters = append(f.Parameters, Parameter{
					identifier: identifierOf(p.Declarator.DirectDeclarator),
					Type:       Type{p.Declarator.Type},
				})
			}
			pList = pList.ParameterList
		}
	}

	return f
}

func emitFunction(f Function) {
	// Function declaration:
	fmt.Printf("func %s(\n", f.GoName())
	for _, p := range f.Parameters {
		fmt.Printf("\t%s %s,\n", p.GoName(), p.Type.GoType())
	}
	if f.ResultType.Kind() == cc.Void {
		fmt.Printf(")")
	} else {
		fmt.Printf(") %s", f.ResultType.GoType())
	}

	// Function body:
	fmt.Printf(" {\n")
	fmt.Printf("\t")
	if f.ResultType.Kind() != cc.Void {
		fmt.Printf("ret := ")
	}
	fmt.Printf("C.%s(\n", f.CName())
	for _, p := range f.Parameters {
		expr := p.GoName()
		if p.Type.RequiresCast() {
			if p.Type.Kind() == cc.Array {
				expr = fmt.Sprintf("(*%s)(&%s[0])", Type{p.Type.Element()}.CGoType(), expr)
			} else {
				expr = fmt.Sprintf("(%s)(%s)", p.Type.CGoType(), expr)
			}
		}
		if p.Type.Kind() == cc.Ptr && !p.Type.IsTypeDef() {
			expr = fmt.Sprintf("unsafe.Pointer(%s)", expr)
		}
		fmt.Printf("\t\t%s,\n", expr)
	}
	fmt.Printf("\t)\n")
	if f.ResultType.Kind() != cc.Void {
		if f.ResultType.RequiresCast() {
			fmt.Printf("\treturn (%s)(ret)\n", f.ResultType.GoType())
		} else {
			fmt.Printf("\treturn ret\n")
		}
	}
	fmt.Printf("}\n")
}

type EnumMember struct {
	identifier string
	Value      interface{}
}

func (m EnumMember) CName() string {
	return m.identifier
}

func (m EnumMember) GoName() string {
	name := m.identifier
	if strings.HasPrefix(name, "VG_") {
		name = name[3:]
	}
	parts := strings.Split(name, "_")
	goName := ""
	for _, p := range parts {
		goName += strings.Title(strings.ToLower(p))
	}
	return goName
}

type Enum struct {
	identifier string
	Members    []EnumMember
}

func (e Enum) CName() string { return e.identifier }

func (e Enum) GoName() string {
	return goName(e.identifier) + "Enum"
}

func (e Enum) CGoName() string {
	return fmt.Sprintf("C.%s", e.identifier)
}

func parseEnum(enDecl *cc.Declarator) Enum {
	//fmt.Println(enDecl)
	constants := enDecl.Type.EnumeratorList()
	e := Enum{
		identifier: identifierOf(enDecl.DirectDeclarator),
		//Type:       enDecl.Type, // TODO: integer base type
		Members: make([]EnumMember, 0, len(constants)),
	}
	for _, m := range constants {

		e.Members = append(e.Members, EnumMember{
			identifier: blessName(m.DefTok.S()),
			Value:      m.Value,
		})
		//m.Declarator.DirectDeclarator
	}
	return e
}

func emitEnum(e Enum) {
	fmt.Printf("type %s int32\n", e.GoName())
	fmt.Printf("const (\n")
	for _, m := range e.Members {
		fmt.Printf("\t%s %s = %v\n", m.GoName(), e.GoName(), m.Value)
	}
	fmt.Printf(")\n")
}

func identifierOf(dd *cc.DirectDeclarator) string {
	switch dd.Case {
	case 0: // IDENTIFIER
		if dd.Token.Val == 0 {
			return ""
		}
		return blessName(dd.Token.S())
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
		return blessName(xc.Dict.S(name))
	} else if rawSpec.IsTypedef() {
		return identifierOf(typ.Declarator().DirectDeclarator)
	}
	return ""
}
