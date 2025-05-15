package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

// loggingMiddleware логирует метод запроса, URI и IP-адрес клиента.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/service/metrics" {
			// Читаем тело запроса
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Ошибка чтения тела запроса: %v", err)
			} else {
				// Восстанавливаем тело запроса для последующих обработчиков
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			// Логируем информацию
			log.Printf("Запрос: %s %s, Body: %s",
				r.Method,
				r.URL.RequestURI(),
				string(bodyBytes))
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
