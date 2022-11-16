package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {

	r := registerRoutes()

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	log.Info("Listening on :", port)
	r.Run(":" + port)
}
