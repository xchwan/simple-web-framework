package main

import (
	"log"
	"net/http"

	"github.com/xchwan/simple-web-framework/framework"
)

// ===== Model =====

// User 代表系統中的使用者。
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ===== 假資料（模擬資料庫）=====

var users = []User{
	{ID: 1, Name: "Alice"},
	{ID: 2, Name: "Bob"},
}

// ===== Web 層業務邏輯 =====

func listUsers(w http.ResponseWriter, r *http.Request) {
	framework.Respond(w, r, http.StatusOK, users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := framework.ParseRequest(r, &newUser); err != nil {
		framework.Respond(w, r, http.StatusBadRequest, nil)
		return
	}
	newUser.ID = len(users) + 1
	users = append(users, newUser)
	framework.Respond(w, r, http.StatusAccepted, nil)
}

func listOrders(w http.ResponseWriter, r *http.Request) {
	framework.Respond(w, r, http.StatusOK, "orders endpoint - coming soon")
}

// ===== 主程式 =====

func main() {
	router := framework.NewRouter()

	router.GET("/api/users", listUsers)
	router.POST("/api/users", createUser)
	router.GET("/api/orders", listOrders)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
