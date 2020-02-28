package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-logger"
	"github.com/pkg/errors"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

func ValidateMsg(m consumer.Message) (content map[string]interface{}, transactionID string, err error) {
	transactionID = m.Headers["X-Request-Id"]
	if transactionID == "" {
		return nil, "", fmt.Errorf("x-request-id not found in kafka message headers. skipping message")
	}

	if err := json.Unmarshal([]byte(m.Body), &content); err != nil {
		return nil, "", errors.Wrap(err, "error: %v - json couldn't be unmarshalled. skipping invalid json")
	}

	return content, transactionID, nil
}

func (h *Notifier) Handle(m consumer.Message) {
	content, transactionID, err := ValidateMsg(m)

	if err != nil {
		logger.Errorf("%v - error consuming message: %v", transactionID, err)
		return
	}

	notificationsReport, err := h.Notify(content, transactionID)
	if err != nil {
		logger.Errorf("error notifying client %+v", err)
		return
	}

	for _, s := range notificationsReport.notifications {
		if s.err != nil {
			logger.WithFields(map[string]interface{}{
				"error": s.err,
			}).Warn("error notifying client")
		} else {
			if s.statusCode != http.StatusOK {
				logger.WithFields(map[string]interface{}{
					"transactionID": s.transactionID,
					"subscription":  s.subscription,
					"statusCode":    s.statusCode,
				}).Warn("invalid status code from webhook")
			}
		}
	}
}
