package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OCRMetrics holds the Prometheus metrics for the OCR service.
type OCRMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorsTotal     *prometheus.CounterVec
}

// NewOCRMetrics creates and registers the OCR metrics.
func NewOCRMetrics() *OCRMetrics {
    // Use a per-instance registry to avoid duplicate registrations during tests
    reg := prometheus.NewRegistry()
    auto := promauto.With(reg)
    return &OCRMetrics{
        RequestsTotal: auto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "ocr_requests_total",
                Help: "Total number of OCR requests.",
            },
            []string{"model"},
        ),
        RequestDuration: auto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "ocr_request_duration_seconds",
                Help:    "Duration of OCR requests.",
                Buckets: prometheus.DefBuckets,
            },
            []string{"model"},
        ),
        ErrorsTotal: auto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "ocr_errors_total",
                Help: "Total number of OCR errors.",
            },
            []string{"model"},
        ),
    }
}
