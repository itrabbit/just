package main

import (
	"github.com/itrabbit/just"
)

func main() {
	app := just.New()
	app.Run("127.0.0.1:8000")
}
