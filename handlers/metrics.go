package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
)

var MessageProcessingTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "message_processing_time_seconds",
		Help:    "Время обработки сообщений, измеренное в секундах, с разделением по типу сообщения.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"message"},
)

//func myFunction() {
//	start := time.Now()
//	// Здесь ваш код функции
//	time.Sleep(100 * time.Millisecond)
//	funcDuration.Observe(time.Since(start).Seconds())
//}
