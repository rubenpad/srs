package main

import (
	"log"

	"github.com/rubenpad/stock-rating-system/cmd/api/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
