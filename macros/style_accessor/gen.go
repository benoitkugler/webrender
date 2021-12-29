package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"

	"github.com/benoitkugler/webrender/css/properties"
)

const (
	OUT_1 = "style/properties/accessors.go"
	OUT_2 = "style/tree/accessors.go"

	TEMPLATE_1 = `
	func (s %[1]s) Get%[2]s() %[3]s {
		return s["%[4]s"].(%[3]s)
	}
	func (s %[1]s) Set%[2]s(v %[3]s) {
		s["%[4]s"] = v
	}

`
	TEMPLATE_2 = `
	func (s *%[1]s) Get%[2]s() %[3]s {
		return s.Get("%[4]s").(%[3]s)
	}
	func (s *%[1]s) Set%[2]s(v %[3]s) {
		s.dict["%[4]s"] = v
	}
	`

	TEMPLATE_ITF = `
    Get%[1]s() %[2]s 
    Set%[1]s(v %[2]s)
	`
)

func main() {
	code_1 := `package properties 
    
	// Code generated from properties/initial_values.go DO NOT EDIT

	`
	code_2 := `package tree 

	// Code generated from properties/initial_values.go DO NOT EDIT

	import pr "github.com/benoitkugler/webrender/css/properties"
	
	`
	code_ITF := "type StyleAccessor interface {"

	var sortedKeys []string
	for k := range properties.InitialValues {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, property := range sortedKeys {
		v := properties.InitialValues[property]

		propertyCamel := camelCase(property)

		typeName := reflect.TypeOf(v).Name()
		// special case for interface values
		if isImage(v) {
			typeName = "Image"
		}

		code_1 += fmt.Sprintf(TEMPLATE_1, "Properties", propertyCamel, typeName, property)
		code_2 += fmt.Sprintf(TEMPLATE_2, "ComputedStyle", propertyCamel, "pr."+typeName, property)
		code_2 += fmt.Sprintf(TEMPLATE_2, "AnonymousStyle", propertyCamel, "pr."+typeName, property)
		code_ITF += fmt.Sprintf(TEMPLATE_ITF, propertyCamel, typeName)
	}

	code_ITF += "}"

	if err := ioutil.WriteFile(OUT_1, []byte(code_1+code_ITF), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(OUT_2, []byte(code_2), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	if err := exec.Command("goimports", "-w", OUT_1).Run(); err != nil {
		log.Fatal(err)
	}
	if err := exec.Command("goimports", "-w", OUT_2).Run(); err != nil {
		log.Fatal(err)
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

func isImage(v interface{}) bool {
	interfaceType := reflect.TypeOf((*properties.Image)(nil)).Elem()
	return reflect.TypeOf(v).Implements(interfaceType)
}
