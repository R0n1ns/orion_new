package middlewares

import (
	"log"
	"net/http"
	"orion/frontclient/services/metrics"
	"orion/frontclient/utils/env"
	"orion/frontclient/utils/jwt"
	"strconv"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/prometheus/client_golang/prometheus"
	"orion/data/manager"
)

// Настроенные глобально лимитеры для публичных и авторизованных запросов
var (
	publicLimiter = tollbooth.NewLimiter(env.RPS_public, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Minute,
	})
	authLimiter = tollbooth.NewLimiter(env.RPS_public, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Minute,
	})
)

func init() {
	// Публичный лимитер — по IP + небольшие всплески
	publicLimiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For"})
	publicLimiter.SetBurst(10)
	publicLimiter.SetMessage("Too many requests. Try again later.")
	publicLimiter.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		metrics.RateLimitBlockedCounter.Inc()
	})

	// Авторизованный лимитер — по ключу (LimitByKeys)
	authLimiter.SetMessage("Too many requests. Try again later.")
	authLimiter.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		metrics.RateLimitBlockedCounter.Inc()
	})
}

// CombinedMiddleware объединяет аутентификацию, rate-limiting и метрики
func CombinedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Публичные эндпоинты
		public := map[string]struct{}{"/service/api/register": {}, "/service/api/login": {},
			"/login": {}, "/register": {}, "/metrics": {}, "/chat": {}, "/": {},
		}
		_, isPublic := public[path]

		var UserId string
		if !isPublic {
			// Аутентификация JWT
			id, err := jwt.ExtractJWT(w, r)
			if err != nil {
				log.Printf("JWT error: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user := manager.GetUserByID(id)
			if user.IsBlocked {
				http.Error(w, "User blocked", http.StatusForbidden)
				return
			}
			UserId = strconv.Itoa(int(id))
		} else {
			// Неавторизованный — позже по IP
			UserId = ""
		}

		// Rate-limit
		if isPublic {
			// Публичный через LimitFuncHandler
			handler := tollbooth.LimitFuncHandler(publicLimiter, next.ServeHTTP)
			handler.ServeHTTP(w, r)
		} else {
			// Авторизованный через LimitByKeys
			if httpErr := tollbooth.LimitByKeys(authLimiter, []string{UserId}); httpErr != nil {
				http.Error(w, httpErr.Message, httpErr.StatusCode)
				return
			}
			next.ServeHTTP(w, r)
		}

		// Метрики и логирование (кроме /metrics)
		if path != "/metrics" {
			if isPublic {
				UserId = "anonymous"
			}
			metrics.GatewayRequestCounter.WithLabelValues(r.Method, path).Inc()
			metrics.UserPathCounter.WithLabelValues(UserId, path).Inc()
			metrics.UserRequestCounter.WithLabelValues(UserId).Inc()

			timer := prometheus.NewTimer(
				metrics.GatewayRequestDuration.WithLabelValues(r.Method, path),
			)
			defer timer.ObserveDuration()

			log.Printf("[%s] %s %s", UserId, r.Method, path)
		}
	})
}
