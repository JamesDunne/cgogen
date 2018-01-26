package main

import (
	"fmt"
	"strings"

	"github.com/cznic/cc"
	"github.com/cznic/xc"
)

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

func goType(typ cc.Type) string {
	switch typ.Kind() {
	case cc.Undefined:
		return "undefined"
	case cc.Void:
		return "void"
	case cc.Ptr:
		return "*" + goType(typ.Element())
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
		return typedefNameOf(typ)
	case cc.TypedefName:
		return typedefNameOf(typ)
	case cc.Function:
		return "func"
	case cc.Array:
		return fmt.Sprintf("[%d]%s", typ.Elements(), goType(typ.Element()))
	default:
		return "???"
	}
}

func main() {
	// Use 64-bit C types model:
	model := models[Arch64]

	// Parse openvg.h main header:
	tu, err := cc.Parse("", []string{"VG/openvg.h"}, model,
		cc.SysIncludePaths([]string{"."}),
		cc.AllowCompatibleTypedefRedefinitions(),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(`package vg

//#include "VG/openvg.h"
import "C"`)

	u := tu
	for u != nil {
		d := u.ExternalDeclaration
		if d.Case != 1 {
			continue
		}

		// Declaration
		decl := d.Declaration
		if decl == nil {
			u = u.TranslationUnit
			continue
		}

		fnDecl := decl.InitDeclaratorListOpt.InitDeclaratorList.InitDeclarator.Declarator
		fn := fnDecl.DirectDeclarator
		if fn.ParameterTypeList == nil {
			fmt.Println(identifierOf(fn))
			u = u.TranslationUnit
			continue
		}

		fnName := identifierOf(fn)
		goFnName := fnName
		if strings.HasPrefix(fnName, "vg") {
			goFnName = fnName[2:]
		}

		// Function:
		fmt.Print("func ", goFnName, "(")

		pList := fn.ParameterTypeList.ParameterList
		if pList.ParameterList == nil && pList.ParameterDeclaration.Declarator == nil {
			// empty void parameter list
			// TODO: check for 'void' type?
			fmt.Print(")")
		} else {
			for pList != nil {
				p := pList.ParameterDeclaration
				//fmt.Println(p)
				if p.Declarator != nil {
					fmt.Print("\n\t", identifierOf(p.Declarator.DirectDeclarator), " ", goType(p.Declarator.Type), ",")
				} else {
					//fmt.Println("\t", p.DeclarationSpecifiers, ",")
				}
				pList = pList.ParameterList
			}
			fmt.Print("\n)")
		}

		resultType := fnDecl.Type.Result()
		if resultType.Kind() == cc.Void {
			fmt.Println()
		} else {
			fmt.Printf(" %s\n", goType(resultType))
		}

		u = u.TranslationUnit
	}
}
