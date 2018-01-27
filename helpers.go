// helpers
package main

var builtinNames = func() map[string]struct{} {
	names := []string{
		"break", "default", "func", "interface", "select", "case", "defer",
		"map", "var", "chan", "else", "goto", "package", "switch", "const",
		"fallthrough", "if", "range", "type", "continue", "for", "import",
		"return", "go", "struct",
	}
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}()

// blessName transforms the name to be a valid name in Go and not a keyword.
func blessName(name []byte) string {
	if _, ok := builtinNames[string(name)]; ok {
		return "_" + string(name)
	}
	return string(name)
}
