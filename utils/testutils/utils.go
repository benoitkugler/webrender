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

func AssertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
}
