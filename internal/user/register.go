package user

import (
	"github.com/xchwan/simple-web-framework/framework"
	"github.com/xchwan/simple-web-framework/framework/scope"
)

// Register 向 router 註冊 user 相關的依賴與路由。
func Register(router *framework.Router) {
	router.Bind("userRepo", func() any {
		return NewUserRepository()
	})
	router.Bind("userService", func() any {
		repo := router.Resolve("userRepo").(*UserRepository)
		return NewUserService(repo)
	}, scope.NewHttpRequestScope())
	router.Bind("userHandler", func() any {
		return NewUserHandler()
	})

	h := router.Resolve("userHandler").(*UserHandler)
	router.POST("/api/users",           h.Register)
	router.POST("/api/users/login",     h.Login)
	router.PATCH("/api/users/{userId}", h.UpdateName)
	router.GET("/api/users",            h.SearchUsers)
}
