package async

import (
	"fmt"
	"testing"
)

func TestResultFrom(t *testing.T) {
	r := ResultFrom[string](getString())
	if r.Value != "hello" {
		t.Errorf("got %v, wanted > %v", r.Value, "hello")
	}
	e := ResultFrom[string](getError())
	if e.Err.Error() != "hello" {
		t.Errorf("got %v, wanted > %v", r.Err, "hello")
	}
}

func getString() (string, error) {
	return "hello", nil
}

func getError() (string, error) {
	return "", fmt.Errorf("hello")
}
