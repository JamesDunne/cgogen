package main

import (
	"fmt"
	"os"
	"strings"

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
		emitEnum(e, o, &VGNamer{})
	}

	for _, f := range functions {
		fmt.Fprintln(o)
		emitFunction(f, o, &VGNamer{})
	}

	return nil
}

type VGNamer struct{}

func (n *VGNamer) EnumName(e Enum) string {
	return goName(e.identifier) + "Enum"
}
func (n *VGNamer) EnumMemberName(m EnumMember) string {
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
func (n *VGNamer) FunctionName(f Function) string {
	goName := f.identifier
	if strings.HasPrefix(goName, "vg") {
		goName = goName[2:]
	}
	return strings.Title(goName)
}
func (n *VGNamer) ParameterName(p Parameter) string {
	return p.identifier
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
