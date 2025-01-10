package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/handlers"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/chats", handlers.GetChatsHandler).Methods("GET")
	r.HandleFunc("/api/chat/{id}", handlers.GetChatHandler).Methods("GET")
	r.HandleFunc("/ws", handlers.WSmanager.HandleWebSocket)

	// Wrap the router with CORS middleware
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Разрешите нужные источники
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
