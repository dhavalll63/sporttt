package utils

import (
	"fmt"
	"math/rand/v2"
)

func GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.IntN(1000000))

}
