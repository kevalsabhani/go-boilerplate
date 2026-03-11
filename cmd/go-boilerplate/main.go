package main

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kevalsabhani/go-boilerplate/internal/config"
)

func main() {
	cfg := config.MustLoad("")
	fmt.Println(cfg)
}
