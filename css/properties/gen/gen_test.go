package main

import (
	"fmt"
	"testing"
)

func TestKebabCase(t *testing.T) {
	fmt.Println(kebabCase("Prop1A"))
}

func TestConstants(t *testing.T) {
	fmt.Println(parseConstants("../properties.go"))
}
