package main

import (
	"fmt"

	"github.com/cznic/cc"
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

	//fmt.Println(tu)

	u := tu
	for u != nil {
		d := u.ExternalDeclaration
		if d.Case == 1 {
			// Declaration
			//fmt.Println(d)
			decl := d.Declaration
			if decl == nil {
				continue
			}

			fnDecl := decl.InitDeclaratorListOpt.InitDeclaratorList.InitDeclarator.Declarator
			fn := fnDecl.DirectDeclarator
			if fn.ParameterTypeList != nil {
				fmt.Print("func ", identifierOf(fn), "(")

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
							fmt.Print("\n\t", identifierOf(p.Declarator.DirectDeclarator), " ", p.Declarator.Type.Kind(), ",")
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
					fmt.Println(" ", resultType.Kind())
				}
			} else {
				fmt.Println(identifierOf(fn))
			}
		}
		u = u.TranslationUnit
	}
}
