package main

import (
	"log"

	"jainam-panchal/nextgen-tranining/gms/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}
