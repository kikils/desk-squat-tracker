package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/kikils/desk-squat-tracker/internal/infrastructure/app"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed all:assets/icon.svg
var icon []byte

func main() {
	if err := app.Run(assets, icon); err != nil {
		log.Fatal(err)
	}
}
