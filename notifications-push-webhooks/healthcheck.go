package main

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/gtg"
)

type HealthCheck struct {
	Consumer consumer.MessageConsumer
}

func NewHealthCheck(kafkaConsumer consumer.MessageConsumer) *HealthCheck {
	return &HealthCheck{
		Consumer: kafkaConsumer,
	}
}

func (h *HealthCheck) Health() func(w http.ResponseWriter, r *http.Request) {
	checks := []fthealth.Check{h.readCheck()}
	hc := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "notifications-push-webhooks",
			Name:        "Notifications Push Webhooks",
			Description: "Checks if all the dependent services are reachable and healthy.",
			Checks:      checks,
		},
		Timeout: 10 * time.Second,
	}
	return fthealth.Handler(hc)
}

func (h *HealthCheck) readCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "read-message-queue-proxy-reachable",
		Name:             "Read Message Queue Proxy Reachable",
		Severity:         2,
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		TechnicalSummary: "Message queue is not reachable/healthy",
		PanicGuide:       "https://dewey.ft.com/up-vm.html",
		Checker:          h.Consumer.ConnectivityCheck,
	}
}

func (h *HealthCheck) GTG() gtg.Status {
	consumerCheck := func() gtg.Status {
		return gtgCheck(h.Consumer.ConnectivityCheck)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		consumerCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
