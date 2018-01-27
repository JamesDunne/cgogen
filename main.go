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

	functions := make([]Function, 0, 50)
	enums := make([]Enum, 0, 50)

	u := tu
	for u != nil {
		e := u.ExternalDeclaration
		if e.Case != 1 {
			continue
		}

		// Declaration
		decl := e.Declaration
		if decl == nil {
			u = u.TranslationUnit
			continue
		}

		d := decl.InitDeclaratorListOpt.InitDeclaratorList.InitDeclarator.Declarator
		dd := d.DirectDeclarator
		if dd.ParameterTypeList != nil {
			f := parseFunction(d)
			functions = append(functions, f)
		} else {
			if d.Type.Kind() == cc.Enum {
				en := parseEnum(d)
				enums = append(enums, en)
			}
		}

		u = u.TranslationUnit
	}

	fmt.Println(`package vg

//#cgo LDFLAGS: -lAmanithVG
//#include "VG/openvg.h"
import "C"

import "unsafe"`)

	for _, e := range enums {
		fmt.Println()
		emitEnum(e)
	}

	for _, f := range functions {
		fmt.Println()
		emitFunction(f)
	}
}
