package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jainam-panchal/nextgen-training/capstone/api"
	"github.com/jainam-panchal/nextgen-training/capstone/pkg/config"
)

func main() {
	cfg := config.Default()
	r := gin.Default()
	api.RegisterRoutes(r)
	if err := r.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
