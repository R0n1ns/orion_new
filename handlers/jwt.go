package handlers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
)

// Claims представляет структуру JWT-токена, содержащую ID пользователя и стандартные поля.
// При необходимости можно расширить эту структуру дополнительными данными.
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// Extract JWT token from cookies
func extractJWT(w http.ResponseWriter, r *http.Request) (uint, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		//http.Error(w, "Unauthorized: Missing JWT", http.StatusUnauthorized)
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
