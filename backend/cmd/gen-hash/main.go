package main

import (
	"fmt"
	"github.com/antclaw/antclaw/internal/auth"
)

func main() {
	hash, err := auth.HashPassword("12345678")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(hash)
}
