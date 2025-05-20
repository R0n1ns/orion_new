package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"orion/frontclient/utils/env"
	"time"
)

// Claims представляет структуру JWT-токена, содержащую ID пользователя и стандартные поля.
// При необходимости можно расширить эту структуру дополнительными данными.
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// Обновлённая extractJWT с проверкой алгоритма и срока действия
func ExtractJWT(w http.ResponseWriter, r *http.Request) (uint, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		return 0, fmt.Errorf("missing token cookie")
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return env.SecretKey, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	// Проверка срока действия (если ExpiresAt задан)
	if claims.ExpiresAt < time.Now().Unix() {
		return 0, fmt.Errorf("token expired")
	}

	return claims.UserID, nil
}
