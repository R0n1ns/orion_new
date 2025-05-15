package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"log"
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
func ExtractJWT(w http.ResponseWriter, r *http.Request) (uint, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		log.Println(1)
		log.Println(err)
		//http.Error(w, "Unauthorized: Missing JWT", http.StatusUnauthorized)
		return 0, err
	}
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		log.Println(1)
		log.Println(err)
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Println(1)
		log.Println(err)
		return 0, err
	}
	//fmt.Println("(claims)", claims.UserID)
	return claims.UserID, nil
}
