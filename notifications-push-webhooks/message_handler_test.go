package main

import (
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
)

const (
	messageTimestamp = "2019-10-02T15:13:26.329Z"
	xRequestID       = "tid_uhsqbhbjdg"
)

var mapper = Notifier{}

func TestTransformMsg_MessageTimestampHeaderMissing(t *testing.T) {
	var message = consumer.Message{
		Headers: map[string]string{
			"X-Request-Id": xRequestID,
		},
		Body: `{
			"id": "3cc23068-e501-11e9-9743-db5a370481bc"
		}`,
	}

	_, _, err := ValidateMsg(message)
	assert.NoError(t, err, "Error not expected when Message-Timestamp header is missing")
	assert.NotEmpty(t, message.Body, "Message body should not be empty")
	assert.NotContains(t, message.Body, "\"lastModified\":", "LastModified field should be generated if header value is missing")
}

func TestTransformMsg_InvalidJson(t *testing.T) {
	var message = consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      xRequestID,
			"Message-Timestamp": messageTimestamp,
		},
		Body: `{{
					"lastModified": "2019-10-02T15:13:19.520Z",
					"publishReference": "tid_uhsqbhbjdg",
					"type": "Article",
					"id": "25781b94-e519-11e9-2686-95f4ddbf4efd"}`,
	}

	_, _, err := ValidateMsg(message)
	assert.Error(t, err, "Expected error when invalid JSON content")
	assert.Contains(t, err.Error(), "json couldn't be unmarshalled. skipping invalid json:", "Expected error message when invalid JSON content")
}

func TestMessage_TransactionIDHeaderMissing(t *testing.T) {
	var message = consumer.Message{
		Headers: map[string]string{
			"Message-Timestamp": messageTimestamp,
		},
		Body: `{}`,
	}

	_, _, err := ValidateMsg(message)
	mapper.Handle(message)
	assert.EqualError(t, err, "x-request-id not found in kafka message headers. skipping message", "Expected error when X-Request-Id is missing")
}
