package gateway

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/prometheus/client_golang/prometheus"
	"orion/data/manager"
	"orion/frontclient/services"
	"orion/frontclient/utils"
)

// Настроенные глобально лимитеры для публичных и авторизованных запросов
var (
	publicLimiter = tollbooth.NewLimiter(utils.RPS_public, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Minute,
	})
	authLimiter = tollbooth.NewLimiter(utils.RPS_public, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Minute,
	})
)

func init() {
	// Публичный лимитер — по IP + небольшие всплески
	publicLimiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For"})
	publicLimiter.SetBurst(10)
	publicLimiter.SetMessage("Too many requests. Try again later.")
	publicLimiter.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		services.RateLimitBlockedCounter.Inc()
	})

	// Авторизованный лимитер — по ключу (LimitByKeys)
	authLimiter.SetMessage("Too many requests. Try again later.")
	authLimiter.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		services.RateLimitBlockedCounter.Inc()
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

		var key string
		if !isPublic {
			// Аутентификация JWT
			id, err := utils.ExtractJWT(w, r)
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
			key = strconv.Itoa(int(id))
		} else {
			// Неавторизованный — позже по IP
			key = ""
		}

		// Rate-limit
		if isPublic {
			// Публичный через LimitFuncHandler
			handler := tollbooth.LimitFuncHandler(publicLimiter, next.ServeHTTP)
			handler.ServeHTTP(w, r)
		} else {
			// Авторизованный через LimitByKeys
			if httpErr := tollbooth.LimitByKeys(authLimiter, []string{key}); httpErr != nil {
				http.Error(w, httpErr.Message, httpErr.StatusCode)
				return
			}
			next.ServeHTTP(w, r)
		}

		// Метрики и логирование (кроме /metrics)
		if path != "/metrics" {
			if isPublic {
				key = "anonymous"
			}
			services.GatewayRequestCounter.WithLabelValues(r.Method, path).Inc()
			services.UserPathCounter.WithLabelValues(key, path).Inc()
			services.UserRequestCounter.WithLabelValues(key).Inc()

			timer := prometheus.NewTimer(
				services.GatewayRequestDuration.WithLabelValues(r.Method, path),
			)
			defer timer.ObserveDuration()

			log.Printf("[%s] %s %s", key, r.Method, path)
		}
	})
}
