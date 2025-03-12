package main

import (
	"log"

	"github.com/rubenpad/srs/cmd/api/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
