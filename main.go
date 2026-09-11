package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/polar-bear-cu/sgt-scheduler/routes"
)

func main() {
	r := gin.Default()
	routes.Register(r)

	log.Println("listening :8090")
	if err := r.Run(":8090"); err != nil {
		log.Fatal(err)
	}
}
