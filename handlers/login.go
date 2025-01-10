package handlers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"orion/data"
	"os"
	"time"
)

// Utility for JSON responses
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// LoginHandler processes login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Некорректный запрос"})
		return
	}

	var user data.User
	if err := data.DB.Where("mail = ?", credentials.Email).First(&user).Error; err != nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Пользователь не найден"})
		return
	}

	// Check password hash
	//if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
	//	jsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Неверный пароль"})
	//	return
	//}
	if credentials.Password != user.Password {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Неверный пароль"})
		return
	}

	if user.IsBlocked {
		jsonResponse(w, http.StatusForbidden, map[string]string{"message": "Пользователь заблокирован"})
		return
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Ошибка генерации токена"})
		return
	}

	// Successful response
	jsonResponse(w, http.StatusOK, map[string]string{"token": tokenString})
}
