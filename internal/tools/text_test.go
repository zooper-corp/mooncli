package tools

import "testing"

func TestToAscii(t *testing.T) {
	a := ToAscii("foo👾bar")
	if a != "foobar" {
		t.Errorf("Expecting 'foobar' got '%v'", a)
	}
}
