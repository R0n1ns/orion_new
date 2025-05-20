package pages

import (
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterRoutes регистрирует маршруты чата на переданном роутере
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/chat", ChatHandler)
	r.HandleFunc("/login", Loginhandler)
	r.HandleFunc("/register", Registerhandler)
	r.HandleFunc("/", ChatHandler)
}

// chatHandler отправляет HTML-страницу chat.html
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "templates/chat.html")
}

// loginhandler отправляет HTML-страницу chat.html
func Loginhandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "templates/login.html")
}

// registerhandler отправляет HTML-страницу chat.html
func Registerhandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "templates/register.html")
}
