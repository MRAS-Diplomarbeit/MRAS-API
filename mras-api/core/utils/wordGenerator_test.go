package utils

import "testing"

func TestGenerator(t *testing.T) {
	test := GenerateCode()
	if len(test) < 10 {
		t.Error()
	}
}
