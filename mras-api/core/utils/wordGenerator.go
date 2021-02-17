package utils

import (
	"github.com/tjarratt/babble"
	"strings"
)

func GenerateCode() string {
	babbler := babble.NewBabbler()
	babbler.Count = 10

	code := babbler.Babble()

	var repl = strings.NewReplacer("'", "")
	code = repl.Replace(code)

	return strings.ToLower(code)
}
