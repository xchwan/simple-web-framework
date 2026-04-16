package main

import (
	"log"

	"github.com/xchwan/simple-web-framework/framework"
	"github.com/xchwan/simple-web-framework/internal/user"
)

func main() {
	router := framework.NewRouter()

	user.Register(router)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
