package gateway

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"orion/frontclient/services/metrics"
	"time"
)

// Исправленный responseRecorder с поддержкой Hijacker
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	hijacked   bool
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if !r.hijacked {
		r.statusCode = statusCode
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	r.hijacked = true
	return hijacker.Hijack()
}

// apiGatewayHandler handles API requests and forwards them to the server
func ApiGatewayHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Record start time
		start := time.Now()

		// Increment request counter
		metrics.GatewayRequestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

		// Create a response recorder to capture the response
		responseRecorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Forward the request to the server
		proxy.ServeHTTP(responseRecorder, r)

		// Record request duration
		duration := time.Since(start).Seconds()
		metrics.GatewayRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		// Log the request
		log.Printf("[API Gateway] %s %s -> %d (%.2fs)", r.Method, r.URL.Path, responseRecorder.statusCode, duration)
	}
}
