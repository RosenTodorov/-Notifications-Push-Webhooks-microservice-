package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
)

type HTTPClientMock struct {
	DoF func(req *http.Request) (*http.Response, error)
}

func (m *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if m.DoF != nil {
		return m.DoF(req)
	}
	return nil, errors.New("not implemented")
}

func TestRequestHandler_Notify(t *testing.T) {

	numHTTPDoCalled := 0
	httpRequests := []*http.Request{}

	tests := []struct {
		name             string
		subscriptions    []Subscription
		httpMock         HTTPClient
		expectError      bool
		numHTTPDoCalled  int
		validateRequests func([]*http.Request) error
	}{
		{
			name: "Success",
			subscriptions: []Subscription{
				{
					EndpointURL:     "http://localhost",
					AuthHeaderKey:   "X-Api-Key",
					AuthHeaderValue: "kulturamiqnko",
				},
				{
					EndpointURL:     "http://localhost",
					AuthHeaderKey:   "X-Api-Key",
					AuthHeaderValue: "kultura",
				},
			},
			httpMock: &HTTPClientMock{
				DoF: func(req *http.Request) (*http.Response, error) {
					numHTTPDoCalled++

					httpRequests = append(httpRequests, req)
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			expectError:     false,
			numHTTPDoCalled: 2,
			validateRequests: func(requests []*http.Request) error {
				assert.Equal(t, 2, len(requests))

				firstRequest := requests[0]
				assert.Equal(t, "POST", firstRequest.Method)

				buf, err := ioutil.ReadAll(firstRequest.Body)
				assert.NoError(t, err)

				assert.Equal(t, `{"Body":"{\"lastModified\":2019-10-02T15:13:19.520Z,\"publishReference\":id_uhsqbhbjdg,\"event\":{\"eventType\":\"test\",\"type\":\"Article\",\"UUID\":\"123\",\"eventTimestamp\":\"0001-01-01T00:00:00Z\"}}","Headers":{"Message-Timestamp":"2019-10-02T15:13:26.329Z","X-Request-Id":"tid_uhsqbhbjdg"}}`, string(buf))

				assert.Equal(t, 200, http.StatusOK)

				return nil
			},
		},
		{
			name:          "If subscriptions is empty",
			subscriptions: []Subscription{},
			httpMock: &HTTPClientMock{
				DoF: func(req *http.Request) (*http.Response, error) {
					numHTTPDoCalled++
					httpRequests = append(httpRequests, req)
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			expectError:     false,
			numHTTPDoCalled: 0,
			validateRequests: func(requests []*http.Request) error {
				assert.Equal(t, 0, len(requests))
				return nil
			},
		},
	}

	var message = consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      xRequestID,
			"Message-Timestamp": messageTimestamp,
		},
		Body: `{"lastModified":2019-10-02T15:13:19.520Z,"publishReference":id_uhsqbhbjdg,"event":{"eventType":"test","type":"Article","UUID":"123","eventTimestamp":"0001-01-01T00:00:00Z"}}`,
	}

	for _, test := range tests {
		fmt.Printf("Runnning test %s \n", test.name)

		// reset
		numHTTPDoCalled = 0
		httpRequests = []*http.Request{}

		notificationsService := Notifier{test.httpMock, test.subscriptions}

		bodyBytes, _ := json.Marshal(message)

		m := make(map[string]interface{})
		error := json.Unmarshal(bodyBytes, &m)
		if error != nil {
			return
		}

		_, err := notificationsService.Notify(m, xRequestID)

		if test.expectError == false {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}

		assert.Equal(t, test.numHTTPDoCalled, numHTTPDoCalled)

		validationError := test.validateRequests(httpRequests)
		assert.NoError(t, validationError)
	}
}
