package handlers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
)

// JWT Claims structure
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// Extract JWT token from cookies
func extractJWT(r *http.Request) (uint, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		fmt.Println(1, err)
		return 0, err
	}
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		fmt.Println(2, err)
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		fmt.Println(3, ok)
		return 0, err
	}
	//fmt.Println("(claims)", claims.UserID)
	return claims.UserID, nil
}
