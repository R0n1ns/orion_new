package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Существующая гистограмма для времени обработки сообщений.
	MessageProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "message_processing_time_seconds",
			Help:    "Время обработки сообщений, измеренное в секундах, с разделением по типу сообщения.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"message"},
	)

	// Счётчик общего количества запросов, разделённый по методу.
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_request_total",
			Help: "Общее количество запросов, обработанных приложением, с разделением по HTTP-методу.",
		},
		[]string{"method"},
	)

	// Гистограмма времени выполнения запросов.
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_duration_seconds",
			Help:    "Время выполнения запросов, измеренное в секундах, с разделением по HTTP-методу.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Счётчик ошибок приложения.
	ErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_error_total",
			Help: "Общее количество ошибок, произошедших в приложении.",
		},
	)

	// Гейдж для отслеживания времени работы приложения.
	AppUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Время работы приложения в секундах.",
		},
	)

	// Гейдж для статической информации о приложении.
	// Значение метрики устанавливается в 1, а информация хранится в лейблах.
	AppInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Информация о приложении (версия, сборка, commit).",
		},
		[]string{"version", "build", "commit"},
	)
)

//func myFunction() {
//	start := time.Now()
//	// Здесь ваш код функции
//	time.Sleep(100 * time.Millisecond)
//	funcDuration.Observe(time.Since(start).Seconds())
//}
