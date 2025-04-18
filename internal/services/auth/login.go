package auth

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"orion/internal/data/managers/postgres"
	"orion/internal/data/models"
	jwt2 "orion/pkg/jwt"
	"os"
	"time"
)

// jsonResponse отправляет ответ клиенту в формате JSON с указанным HTTP-статусом.
// При возникновении ошибок кодирования JSON они игнорируются, так как отправка заголовков уже завершена.
//
// Параметры:
//   - w: http.ResponseWriter для отправки ответа.
//   - status: HTTP-статус ответа.
//   - data: данные для кодирования в JSON.
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Обработка ошибки при кодировании JSON. Если возникнет ошибка, она выводится в лог.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Поскольку заголовки уже отправлены, повторно изменить статус невозможно.
		// Логируем ошибку для последующего анализа.
		// Можно дополнительно настроить внешние уведомления об ошибке.
		// Пример: log.Printf("failed to encode json response: %v", err)
	}
}

// LoginHandler обрабатывает запросы на вход пользователя.
// Функция выполняет следующие шаги:
//  1. Декодирует входящий JSON-запрос и проверяет корректность полученных данных.
//  2. Ищет пользователя по email в базе данных.
//  3. Сравнивает пароль из запроса с сохранённым в базе (здесь используется простое сравнение,
//     для реального проекта рекомендуется использовать bcrypt или другую надёжную хеширующую функцию).
//  4. Проверяет, не заблокирован ли пользователь.
//  5. Создаёт JWT-токен с временем жизни 24 часа.
//  6. Отправляет сгенерированный токен клиенту.
//
// В случае возникновения ошибки на любом этапе возвращается соответствующий HTTP-статус и сообщение об ошибке.
//
// Обработка ошибок:
//   - Если JSON-запрос невозможно декодировать, возвращается HTTP 400 с сообщением "Некорректный запрос".
//   - Если пользователь не найден по email, возвращается HTTP 401 с сообщением "Пользователь не найден".
//   - Если пароль неверный, возвращается HTTP 401 с сообщением "Неверный пароль".
//   - Если пользователь заблокирован, возвращается HTTP 403 с сообщением "Пользователь заблокирован".
//   - Если при создании JWT-токена происходит ошибка, возвращается HTTP 500 с сообщением "Ошибка генерации токена".
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Структура для декодирования входящих данных
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Декодирование JSON-запроса. В случае ошибки возвращаем HTTP 400.
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Некорректный запрос"})
		return
	}

	// Поиск пользователя по email в базе данных.
	var user models.User
	if err := postgres.DB.Where("mail = ?", credentials.Email).First(&user).Error; err != nil {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Пользователь не найден"})
		return
	}

	// Проверка пароля.
	// В данном примере используется простое сравнение, но для реальных приложений рекомендуется использовать bcrypt.
	if credentials.Password != user.Password {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Неверный пароль"})
		return
	}

	// Проверка, заблокирован ли пользователь.
	if user.IsBlocked {
		jsonResponse(w, http.StatusForbidden, map[string]string{"message": "Пользователь заблокирован"})
		return
	}

	// Создание JWT-токена с временем жизни 24 часа.
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt2.Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Создаём токен с использованием алгоритма HS256 и наших claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	// Генерация строкового представления токена.
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// Если ошибка возникает при генерации токена, возвращаем HTTP 500.
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Ошибка генерации токена"})
		return
	}

	// Отправка успешного ответа с токеном.
	jsonResponse(w, http.StatusOK, map[string]string{"token": tokenString})
}
