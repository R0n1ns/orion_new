package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Регистрируем метрику в реестре Prometheus
	prometheus.MustRegister(MessageProcessingTime)
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ErrorCounter)
	prometheus.MustRegister(AppUptime)
	prometheus.MustRegister(AppInfo)
	prometheus.MustRegister(ActiveChatsGauge)
}

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
	// Определяем метрику типа Gauge для отслеживания количества активных чатов
	ActiveChatsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ws_manager_active_chats_total",
		Help: "Количество активных чатов в ws manager",
	})
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
