package main

import (
	"exdex/internal/app"
	"exdex/internal/router"
)

func main() {

	r := router.NewRouter()
	ap := app.NewApp(r)
	ap.Start()
}
