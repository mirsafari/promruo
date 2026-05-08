package main

import (
	"fmt"
	"os"

	"github.com/mirsafari/promruo/configs"
)

func main() {
	cfg := configs.NewConfig()

	fmt.Fprintf(os.Stdout, "ENABLE_WEB=%t\n", cfg.EnableWeb)
}
