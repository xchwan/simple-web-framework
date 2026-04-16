package main

import (
	"log"

	"github.com/xchwan/simple-web-framework/framework"
	"github.com/xchwan/simple-web-framework/internal/user"
)

func main() {
	userRepo    := user.NewUserRepository()
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	router := framework.NewRouter()

	// 靜態路由（/api/users/login）必須在 wildcard 路由（/api/users/{userId}）之前註冊
	router.POST("/api/users",           userHandler.Register)
	router.POST("/api/users/login",     userHandler.Login)
	router.PATCH("/api/users/{userId}", userHandler.UpdateName)
	router.GET("/api/users",            userHandler.SearchUsers)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
