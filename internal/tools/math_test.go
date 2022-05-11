package tools

import "testing"

func TestMin(t *testing.T) {
	a := Min(123, 13)
	if a != 13 {
		t.Errorf("Expected '13' got '%v'", a)
	}
}

func TestHumanize(t *testing.T) {
	h := Humanize(123416789.123)
	if h != "123.4M" {
		t.Errorf("Expected '123.4M' got '%v'", h)
	}
	h = Humanize(123416.123)
	if h != "123.4K" {
		t.Errorf("Expected '123.4K' got '%v'", h)
	}
	h = Humanize(123.123)
	if h != "123.1" {
		t.Errorf("Expected '123.1' got '%v'", h)
	}
	h = Humanize(0.123)
	if h != "0.123" {
		t.Errorf("Expected '0.123' got '%v'", h)
	}
}
