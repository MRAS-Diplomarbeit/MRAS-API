package utils

import "github.com/tjarratt/babble"

func GenerateCode() string {
	babbler := babble.NewBabbler()
	babbler.Count = 10
	return babbler.Babble()
}
