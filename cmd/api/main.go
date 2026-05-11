package main

import (
	"fmt"
	"os"

	"github.com/faqihyugos/coffee-pos/config"
)

func main() {
	fmt.Println("Coffee Shop POS starting...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Config loaded. Starting server on port :" + cfg.AppPort)
}
