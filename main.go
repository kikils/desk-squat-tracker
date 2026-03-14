package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/kikils/desk-squat-tracker/internal/infrastructure/app"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/standup.png
var iconStandup []byte

//go:embed assets/squat.png
var iconSquat []byte

func main() {
	if err := app.Run(assets, iconStandup, iconSquat); err != nil {
		log.Fatal(err)
	}
}
