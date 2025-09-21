package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateID generates a unique memory ID using timestamp and random suffix
func GenerateID() string {
	timestamp := time.Now().Unix()
	random := rand.Intn(999999)
	return fmt.Sprintf("mem_%x_%06x", timestamp, random)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
