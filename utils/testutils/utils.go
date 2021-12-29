package testutils

import (
	"reflect"
	"testing"
)

func AssertEqual(t *testing.T, got, exp interface{}, context string) {
	t.Helper()
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("%s: expected\n%v\n got \n%v", context, exp, got)
	}
}
