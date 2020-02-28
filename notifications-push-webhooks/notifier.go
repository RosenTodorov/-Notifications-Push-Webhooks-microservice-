package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type Notifier struct {
	httpClient    HTTPClient
	subscriptions []Subscription
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Notification struct {
	err           error
	subscription  Subscription
	statusCode    int
	transactionID string
}
type NotificationReport struct {
	notifications []*Notification
}

func (n *Notifier) Notify(content map[string]interface{}, transactionID string) (*NotificationReport, error) {
	bodyBytes, err := json.Marshal(content)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling content")
	}

	if len(n.subscriptions) == 0 {
		return &NotificationReport{}, nil
	}

	nr := &NotificationReport{}

	for _, s := range n.subscriptions {
		var body = bytes.NewReader(bodyBytes)

		req, err := http.NewRequest("POST", s.EndpointURL, body)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			nr.notifications = append(nr.notifications, &Notification{
				err:           errors.Wrap(err, "Error notifying client"),
				subscription:  s,
				transactionID: transactionID,
			})
			continue
		}

		if s.AuthHeaderKey != "" && s.AuthHeaderValue != "" {
			req.Header.Set(s.AuthHeaderKey, s.AuthHeaderValue)
		}

		response, err := n.httpClient.Do(req)
		if err != nil {
			nr.notifications = append(nr.notifications, &Notification{
				err:           errors.Wrap(err, "Error making http request"),
				subscription:  s,
				transactionID: transactionID,
			})
			continue
		}

		nr.notifications = append(nr.notifications, &Notification{
			subscription: s,
			statusCode:   response.StatusCode,
		})

		if response.Body != nil {
			response.Body.Close()
		}
	}
	return nr, nil
}
