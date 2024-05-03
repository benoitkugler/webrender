package testutils

import (
	"reflect"
	"testing"
)

func AssertEqual(t *testing.T, got, exp interface{}) {
	t.Helper()
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected\n%v\n got \n%v", exp, got)
	}
}
