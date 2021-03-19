package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateMulticastIP() string {

	rand.Seed(time.Now().UnixNano())

	max := 255

	rand1 := rand.Intn(max)
	rand2 := rand.Intn(max)
	rand3 := rand.Intn(max)

	return fmt.Sprintf("239.%d.%d.%d", rand1, rand2, rand3)
}
