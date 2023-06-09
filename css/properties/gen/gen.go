package main

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/benoitkugler/webrender/css/properties"
)

const (
	OUT_1 = "props_gen.go"
	OUT_2 = "../../html/tree/accessors.go"

	TEMPLATE_1 = `
	func (s %[1]s) Get%[2]s() %[3]s { return s[%[4]s].(%[3]s)	}
	func (s %[1]s) Set%[2]s(v %[3]s) { s[%[4]s] = v }

`
	TEMPLATE_2 = `
	func (s *%[1]s) Get%[2]s() %[3]s {
		return s.Get(pr.%[4]s.Key()).(%[3]s)
	}
	func (s *%[1]s) Set%[2]s(v %[3]s) {
		s.propsCache.known[pr.%[4]s] = v
	}
	`

	TEMPLATE_ITF = `
    Get%[1]s() %[2]s 
    Set%[1]s(v %[2]s)
	`
)

func main() {
	code_1 := `package properties 
    
	// Code generated from properties/properties.go DO NOT EDIT

	`
	code_2 := `package tree 

	// Code generated from properties/properties.go DO NOT EDIT

	import pr "github.com/benoitkugler/webrender/css/properties"
	
	`
	code_ITF := "type StyleAccessor interface {"

	code_strings := `var propsNames = [...]string{
		`
	code_strings_rev := `
	// PropsFromNames maps CSS property names to internal enum tags.
	var PropsFromNames = map[string]KnownProp{
		`

	props := parseConstants("properties.go")
	sort.Slice(props, func(i, j int) bool { return props[i].propName < props[j].propName })

	for _, item := range props {
		property := item.value
		v := properties.InitialValues[property]

		propertyCamel := item.varName[1:]

		typeName := reflect.TypeOf(v).Name()
		// special case for interface values
		if isImage(v) {
			typeName = "Image"
		}

		code_1 += fmt.Sprintf(TEMPLATE_1, "Properties", propertyCamel, typeName, item.varName)
		code_2 += fmt.Sprintf(TEMPLATE_2, "ComputedStyle", propertyCamel, "pr."+typeName, item.varName)
		code_2 += fmt.Sprintf(TEMPLATE_2, "AnonymousStyle", propertyCamel, "pr."+typeName, item.varName)
		code_ITF += fmt.Sprintf(TEMPLATE_ITF, propertyCamel, typeName)
		code_strings += fmt.Sprintf("%s: %q,\n", item.varName, item.propName)
		code_strings_rev += fmt.Sprintf("%q: %s,\n", item.propName, item.varName)
	}

	code_ITF += "}\n"
	code_strings += "}\n"
	code_strings_rev += "}\n"

	if err := os.WriteFile(OUT_1, []byte(code_1+code_ITF+code_strings+code_strings_rev), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.WriteFile(OUT_2, []byte(code_2), os.ModePerm); err != nil {
		panic(err)
	}

	if err := exec.Command("goimports", "-w", OUT_1).Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("goimports", "-w", OUT_2).Run(); err != nil {
		panic(err)
	}
	fmt.Println("Generated", OUT_1, OUT_2)
}

func camelCase(s string) string {
	out := ""
	for _, part := range strings.Split(s, "_") {
		out += strings.Title(part)
	}
	return out
}

func kebabCase(s string) string {
	var out strings.Builder
	for i, r := range s {
		if i != 0 && unicode.IsUpper(r) {
			out.WriteRune('-')
		}
		out.WriteRune(unicode.ToLower(r))
	}
	return out.String()
}

func isImage(v interface{}) bool {
	interfaceType := reflect.TypeOf((*properties.Image)(nil)).Elem()
	return reflect.TypeOf(v).Implements(interfaceType)
}

type prop struct {
	value    properties.KnownProp
	varName  string
	propName string // in CSS form
}

func parseConstants(fn string) (out []prop) {
	b, err := os.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	inEnum := false
	var val properties.KnownProp
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "const") {
			inEnum = true
			continue
		}

		if inEnum && strings.HasPrefix(line, "P") {
			val++
			varName := line
			propName := kebabCase(line[1:])
			out = append(out, prop{val, varName, propName})
		}

		if inEnum && strings.HasPrefix(line, ")") {
			inEnum = false
		}
	}
	return out
}
