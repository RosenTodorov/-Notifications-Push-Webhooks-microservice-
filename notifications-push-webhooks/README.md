# notifications-push-webhooks

[![Circle CI](https://circleci.com/gh/Financial-Times/notifications-push-webhooks/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/notifications-push-webhooks/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/notifications-push-webhooks)](https://goreportcard.com/report/github.com/Financial-Times/notifications-push-webhooks) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/notifications-push-webhooks/badge.svg)](https://coveralls.io/github/Financial-Times/notifications-push-webhooks)

## Introduction

It is a microservice that provides push notifications and consumes a specific Apache Kafka topic group, then it pushes a notification for each member

## Installation
      
In order to install, execute the following steps:

Download the source code, dependencies and test dependencies:

        curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
        go get -u github.com/Financial-Times/notifications-push-webhooks
        cd $GOPATH/src/github.com/Financial-Times/notifications-push-webhooks
        dep ensure
        go build .

## Running locally

1. Cluster setup, start the Kafka system (upp-managed-kafka-poc) and start the "notifications-push-webhooks" service.
- https://github.com/Financial-Times/upp-managed-kafka-poc

Starting and stopping the Apache Kafka system is facilitated by a simple bash script:
        
        # In one terminal go to these repo's folder and start a Kafka cluster using the init script
        ./local-setup/upp-kafka.sh cluster start

        # In another terminal start the "notifications-push-webhooks" service
        notifications-push-webhooks --queue-addresses=http://localhost:8082 --group=notificationsPushWebhooks --read-topic=PostPublicationEvents --subscriptions="[{\"url\":\"http://localhost:8888/client\", \"authHeaderKey\":\"X-Api-Key\", \"authHeaderValue\":\"kulturamiqnko\"},{\"url\":\"http://localhost:8888/client\", \"authHeaderKey\":\"X-Api-Key\", \"authHeaderValue\":\"kultura\"}]"

2. We need to post a sample message to the "PostPublicationEvents" Kafka topic:
      
       # In another terminal post a sample message to the "PostPublicationEvents"
       kafkacat -P -b localhost:9092 -t PostPublicationEvents -p 0 PostPublicationEvents.message

## Expected behaviour

The mapper reads events from the PostPublicationEvents topic and then it pushes a notification for each member.

- Create a simple service to receive the POST request for each member.
 ```
 package main
 
 import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
 ) 
 
 func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	b, _ := ioutil.ReadAll(r.Body)

	ua := r.Header.Get("X-Api-Key")

	fmt.Printf("URL %s!\nBody: %s\nHeader: %s", r.URL.Path[1:], b, ua)
 }
 
 func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8888", nil))
 }
```
- In another terminal start the service with it's name.

You should receive a response body like:

        Body: {"contentUri":"http://methode-article-mapper.svc.ft.com/content/3cc23068-e501-11e9-9743-db5a370481bc","lastModified":"2019-10-02T15:13:19.52Z","payload":{"accessLevel":"subscribed","alternativeStandfirsts":{"promotionalStandfirst":null},"alternativeTitles":{"contentPackageTitle":null,"promotionalTitle":"Lebanon eases dollar flow for importers as crisis grows"},"body":"\u003cbody\u003e\u003ccontent data-embedded=\"true\" id=\"25781b94-e519-11e9-2686-95f4ddbf4efd\" type=\"http://www.ft.com/ontology/content/ImageSet\"\u003e\u003c/content\u003e\u003cp\u003eLebanon’s central bank has made it easier for some importers to access dollars, as the country grapples with an economic and monetary crisis that has seen protesters take to the streets.\u003c/p\u003e\n\u003cp\u003eThe \u003ccontent id=\"fd499728-e474-11e9-9743-db5a370481bc\" type=\"http://www.ft.com/ontology/content/Article\"\u003ebank\u003c/content\u003e late on Tuesday introduced a foreign exchange measure to help ensure imports of essential goods, amid fears that the growing squeeze on dollars in one of the world’s most indebted countries could cut deliveries of wheat and fuel.\u003c/p\u003e\n\u003cp\u003eIt was another sign of the extreme pressure on Lebanon’s monetary system, as the import-dependent country struggles to cope with a deepening trade deficit and a slowing economy. Prime Minister Sa’ad Hariri has declared an economic state of emergency. \u003c/p\u003e\n\u003cp\u003e“It is true that we are going through difficult economic conditions, but we can overcome them,” said Mr Hariri, adding that “we are now working on the issue of chaos among some exchange offices and this issue is on track”.\u003c/p\u003e\n\u003cp\u003eSami Atallah, director at the Lebanese Center for Policy Studies think-tank in Beirut, said: “We’ve basically been living beyond our means and importing as if there’s no tomorrow. \u003c/p\u003e\n\u003cp\u003e“We need to rebalance the economy.”\u003c/p\u003e\n\u003cp\u003eBut frequent deadlock within Lebanon’s power-sharing confessional government has undermined efforts at structural reform, and debt has ballooned to more than 150 per cent of gross domestic product — the world’s third-highest such ratio.\u003c/p\u003e\n\u003cp\u003eAs the crisis has escalated, demonstrators took to the streets of Beirut this weekend to protest joblessness and corruption, raising the prospect of further instability in a country that has the world’s biggest population of refugees per capita.\u003c/p\u003e\n\u003cp\u003eAt the same time, business owners have complained that banks are increasingly unwilling to convert Lebanese lira into dollars and that market demand for dollars is driving up the cost of converting at more informal foreign exchange brokers.\u003c/p\u003e\n\u003cp\u003eSyndicates representing both millers and gas station owners last week warned that an absence of dollars made it difficult for them to pay their suppliers in hard currency. In response, the Banque du Liban (BdL) announced on Tuesday that it would guarantee commercial banks had a supply of dollars for importers of wheat, oil derivatives and medicine at the official exchange rate of 1,507 lira to the US dollar.\u003c/p\u003e\n\u003cp\u003eThis effectively creates a dual exchange rate where importers of essential goods pay less for US dollars than other traders. One small high-end alcohol importer said that the rising price of converting lira into the dollars he needs to pay his overseas suppliers would boost prices. “If we get paid in lira, we’re going to get screwed,” he said.\u003c/p\u003e\n\u003cp\u003eHigh interest rates had traditionally attracted dollar deposits from Lebanon’s huge diaspora to meet its foreign currency needs but deposits fell in May for the first time in decades — according to Goldman Sachs — interpreted as a sign of waning confidence in the economy.\u003c/p\u003e\n\u003cp\u003eRating agency Moody’s on Tuesday warned that government reliance on BdL reserves to pay off forthcoming debt obligations “risks destabilising the BdL’s ability to sustain the currency peg”. \u003c/p\u003e\n\u003cp\u003eBdL governor Riad Salame has insisted that the BdL’s shrinking foreign exchange reserves — down 9 per cent year on year to $31bn as of July 2019 — are not a cause for concern.\u003c/p\u003e\n\u003cp\u003eMoody’s also announced that Lebanon was under review for a downgrade, weeks after Fitch cut the sovereign’s rating, pushing it deeper into junk territory.\u003c/p\u003e\n\n\u003c/body\u003e","brands":[{"id":"http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}],"byline":"Chloe Cornish in Beirut","canBeDistributed":"yes","canBeSyndicated":"yes","canonicalWebUrl":"https://www.ft.com/content/3cc23068-e501-11e9-9743-db5a370481bc","comments":{"enabled":true},"contentPackage":null,"copyright":null,"description":null,"editorialDesk":"/FT/WorldNews/Middle East and North Africa","externalBinaryUrl":null,"firstPublishedDate":"2019-10-02T13:49:30.000Z","identifiers":[{"authority":"http://api.ft.com/system/FTCOM-METHODE","identifierValue":"3cc23068-e501-11e9-9743-db5a370481bc"}],"internalAnalyticsTags":null,"internalBinaryUrl":null,"lastModified":"2019-10-02T15:13:19.520Z","mainImage":"25781b94-e519-11e9-2686-95f4ddbf4efd","masterSource":null,"mediaType":null,"members":null,"pixelHeight":null,"pixelWidth":null,"publishReference":"tid_uhsqbhbjdg","publishedDate":"2019-10-02T13:49:30.000Z","rightsGroup":null,"standfirst":"Shortage of hard currency raises fears about world’s most-indebted economy","standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"storyPackage":null,"title":"Lebanon eases dollar flow for importers as crisis grows","type":"Article","uuid":"3cc23068-e501-11e9-9743-db5a370481bc","webUrl":"https://www.ft.com/content/3cc23068-e501-11e9-9743-db5a370481bc"}}
        Header: kultura

## Build and deployment

* Built by Docker Hub on merge to master: [notifications-push-webhooks](https://hub.docker.com/coco/notifications-push-webhooks/)
* CI provided by CircleCI: [notifications-push-webhooks](https://circleci.com/gh/Financial-Times/notifications-push-webhooks)

## Healthchecks

Admin endpoints are:

/__gtg
/__health
/__build-info

## Logging

* The application uses go-logger; the log file is initialised in main.go.

Task:
Implement push notifications with webhooks

Create project with cookiecutter 'notifications-push-webhooks':
-cookiecutter project
-helm config
-jenkins config
-setup CircleCI
Implement reading messages from Kafka via REST proxy - Read messages from Kafka on PostCMSPublicationEvent in separate consumer group.
Implement HTTP client that notifies clients about new publication - no authentication, Standard HTTP POST request
Implement loading configuration about webhook addresses and authentication - create custom secret in k8s via helm for this service.
Inside the secret we will have JSON in the following format:
[
{"url":"http://example.com", "authHeaderKey":"", "authHeaderValue":""},
{"url":"http://example.com", "authHeaderKey":"X-Api-Key", "authHeaderValue":"kulturamiqnko"},
]