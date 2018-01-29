package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cznic/cc"
)

func generateCgo(srcPaths []string, packageName string, outPath string, namer Namer) error {
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
			if !namer.IgnoreFunction(f.identifier) {
				functions = append(functions, f)
			}
		} else {
			if d.Type.Kind() == cc.Enum {
				en := parseEnum(d)
				if !namer.IgnoreEnum(en.identifier) {
					enums = append(enums, en)
				}
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
		emitEnum(e, o, namer)
	}

	for _, f := range functions {
		fmt.Fprintln(o)
		emitFunction(f, o, namer)
	}

	return nil
}

func goName(name string) string {
	if strings.HasPrefix(name, "VG") {
		return Export(name[2:])
	}
	return Export(name)
}

type VGNamer struct{}

func (n *VGNamer) IgnoreEnum(name string) bool {
	return false
}
func (n *VGNamer) IgnoreFunction(name string) bool {
	return false
}
func (n *VGNamer) EnumName(e Enum) string {
	name := e.identifier
	if strings.HasPrefix(name, "VG") {
		name = name[2:]
	}
	return Export(name) + "Enum"
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

type VGUNamer struct{}

func (n *VGUNamer) IgnoreEnum(name string) bool {
	return !strings.HasPrefix(name, "VGU")
}
func (n *VGUNamer) IgnoreFunction(name string) bool {
	return !strings.HasPrefix(name, "vgu")
}
func (n *VGUNamer) EnumName(e Enum) string {
	name := e.identifier
	if strings.HasPrefix(name, "VGU") {
		name = name[3:]
	}
	return Export(name) + "Enum"
}
func (n *VGUNamer) EnumMemberName(m EnumMember) string {
	name := m.identifier
	if strings.HasPrefix(name, "VGU_") {
		name = name[4:]
	}
	parts := strings.Split(name, "_")
	goName := ""
	for _, p := range parts {
		goName += strings.Title(strings.ToLower(p))
	}
	return goName
}
func (n *VGUNamer) FunctionName(f Function) string {
	goName := f.identifier
	if strings.HasPrefix(goName, "vgu") {
		goName = goName[3:]
	}
	return strings.Title(goName)
}
func (n *VGUNamer) ParameterName(p Parameter) string {
	return p.identifier
}

func main() {
	var err error
	err = generateCgo([]string{"VG/openvg.h"}, "vg", "../golang-openvg/vg/vg.go", &VGNamer{})
	if err != nil {
		panic(err)
	}
	err = generateCgo([]string{"VG/vgu.h"}, "vgu", "../golang-openvg/vgu/vgu.go", &VGUNamer{})
	if err != nil {
		panic(err)
	}
}
