package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
)

type mockConsumerInstance struct {
	isConnectionHealthy bool
}

func (consumer *mockConsumerInstance) ConnectivityCheck() (string, error) {
	if consumer.isConnectionHealthy {
		return "", nil
	}

	return "", errors.New("error connecting to the queue")
}

func (consumer *mockConsumerInstance) Start() {
}

func (consumer *mockConsumerInstance) Stop() {
}

func initializeHealthCheck(isConsumerConnectionHealthy bool) *HealthCheck {
	return &HealthCheck{
		Consumer: &mockConsumerInstance{isConnectionHealthy: isConsumerConnectionHealthy},
	}
}

func TestNewHealthCheck(t *testing.T) {
	healthCheck := NewHealthCheck(
		consumer.NewConsumer(consumer.QueueConfig{}, func(m consumer.Message) {}, http.DefaultClient),
	)

	assert.NotNil(t, healthCheck.Consumer)
}

func TestHappyHealthCheck(t *testing.T) {
	healthCheck := initializeHealthCheck(true)

	request := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	healthCheck.Health()(w, request)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":true`, "Read message queue proxy healthcheck should be happy")
}

func TestUnhappyHealthCheck(t *testing.T) {
	healthCheck := initializeHealthCheck(false)

	request := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	healthCheck.Health()(w, request)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":false`, "Read message queue proxy healthcheck should be unhappy")
}

func TestGTGHappyFlow(t *testing.T) {
	healthCheck := initializeHealthCheck(true)

	status := healthCheck.GTG()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGTGBrokenConsumer(t *testing.T) {
	healthCheck := initializeHealthCheck(false)

	status := healthCheck.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "error connecting to the queue", status.Message)
}
