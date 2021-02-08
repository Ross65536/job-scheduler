package testutil

import (
	"reflect"
	"strings"
	"testing"
)

func AssertNotError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func AssertContains(t *testing.T, actual, substr string) {
	if !strings.Contains(actual, substr) {
		t.Fatalf("String %s doesn't contain %s", actual, substr)
	}
}

func AssertEquals(t *testing.T, actual interface{}, expected interface{}) {

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Invalid field, expected %s, was %s", expected, actual)
	}
}

func AssertNotEquals(t *testing.T, actual interface{}, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		t.Fatalf("Invalid field, expected %s to be different from %s", expected, actual)
	}
}
