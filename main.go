package main

import (
	"fmt"
	"os"

	"github.com/cznic/cc"
)

func generateCgo(srcPaths []string, packageName string, outPath string) error {
	// Use 64-bit C types model:
	model := &cc.Model{
		Items: make(map[cc.Kind]cc.ModelItem),
	}
	for k, v := range models[Arch64].Items {
		model.Items[k] = v
	}

	// Parse openvg.h main header:
	tu, err := cc.Parse("", srcPaths, model,
		cc.SysIncludePaths([]string{"."}),
		cc.AllowCompatibleTypedefRedefinitions(),
	)
	if err != nil {
		return err
	}

	o, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer o.Close()

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

	fmt.Fprintf(o, `package %s

//#cgo LDFLAGS: -lAmanithVG
`, packageName)
	for _, s := range srcPaths {
		fmt.Fprintf(o, "//#include \"%s\"\n", s)
	}
	fmt.Fprintln(o, `import "C"

import "unsafe"`)

	for _, e := range enums {
		fmt.Fprintln(o)
		emitEnum(e, o)
	}

	for _, f := range functions {
		fmt.Fprintln(o)
		emitFunction(f, o)
	}

	return nil
}

func main() {
	var err error
	err = generateCgo([]string{"VG/openvg.h"}, "vg", "../golang-openvg/vg/vg.go")
	if err != nil {
		panic(err)
	}
	//	err = generateCgo([]string{"VG/vgu.h"}, "vgu", "../golang-openvg/vgu/vgu.go")
	//	if err != nil {
	//		panic(err)
	//	}
}
