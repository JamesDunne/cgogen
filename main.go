package main

import (
	"fmt"

	"github.com/cznic/cc"
)

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
			fmt.Printf("return ")
		}
		fmt.Printf("C.%s(\n", f.CName())
		for _, p := range f.Parameters {
			if p.Type.RequiresCast() {
				fmt.Printf("\t\t(%s)(%s),\n", p.Type.CGoType(), p.GoName())
			} else {
				fmt.Printf("\t\t%s,\n", p.GoName())
			}
		}
		fmt.Printf("\t)\n")
		fmt.Printf("}\n")

		u = u.TranslationUnit
	}
}
