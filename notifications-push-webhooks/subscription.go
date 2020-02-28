package main

type Subscription struct {
	EndpointURL     string `json:"url"`
	AuthHeaderKey   string `json:"authHeaderKey,omitempty"`
	AuthHeaderValue string `json:"authHeaderValue,omitempty"`
}

type Subscriptions []Subscription
