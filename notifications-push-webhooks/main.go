package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"

	"github.com/Financial-Times/go-logger"
	log "github.com/Financial-Times/go-logger"
	cli "github.com/jawher/mow.cli"

	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"

	status "github.com/Financial-Times/service-status-go/httphandlers"
)

const (
	serviceName    = "notifications-push-webhooks"
	appDescription = "It is a microservice that provides push notifications and consumes a specific Apache Kafka topic group, then it pushes a notification for each member"
)

func main() {
	app := cli.App(serviceName, appDescription)

	addresses := app.Strings(cli.StringsOpt{
		Name:   "queue-addresses",
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "Q_ADDR",
	})

	group := app.String(cli.StringOpt{
		Name:   "group",
		Desc:   "Group used to read the messages from the queue.",
		EnvVar: "Q_GROUP",
	})

	readTopic := app.String(cli.StringOpt{
		Name:   "read-topic",
		Desc:   "The topic to read the meassages from.",
		EnvVar: "Q_READ_TOPIC",
	})

	readQueue := app.String(cli.StringOpt{
		Name:   "read-queue",
		Desc:   "The queue to read the meassages from.",
		EnvVar: "Q_READ_QUEUE",
	})

	subscriptions := app.String(cli.StringOpt{
		Name:   "subscriptions",
		Desc:   "subscription key and value.",
		EnvVar: "Q_SUBSCRIPTION",
	})

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "notifications-push-webhooks-system-code",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})

	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "notifications-push-webhooks",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})

	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "INFO",
		Desc:   "Logging level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "LOG_LEVEL",
	})

	app.Action = func() {
		log.InitDefaultLogger(serviceName)

		log.InitLogger(*appSystemCode, *logLevel)

		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		log.Infof("[Startup] notifications-push-webhooks is starting ")

		if len(*addresses) == 0 {
			logger.Error("No queue address provided. Quitting...")
			cli.Exit(1)
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConnsPerHost:   20,
				TLSHandshakeTimeout:   3 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		consumerConfig := consumer.QueueConfig{
			Addrs:                *addresses,
			Group:                *group,
			Topic:                *readTopic,
			Queue:                *readQueue,
			ConcurrentProcessing: false,
			AutoCommitEnable:     true,
		}

		var subscriptionConfig Subscriptions
		var err = json.Unmarshal([]byte(*subscriptions), &subscriptionConfig)
		if err != nil {
			log.Errorf("Error: %v - JSON couldn't be unmarshalled", err)
			return
		}

		var rh = Notifier{httpClient, subscriptionConfig}
		go func() {
			serveEndpoints(*appSystemCode, *appName, *port, Notifier{})
		}()

		var messageConsumer = consumer.NewConsumer(consumerConfig, rh.Handle, httpClient)
		consumeUntilSigterm(messageConsumer)

		waitForSignal()
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func consumeUntilSigterm(messageConsumer consumer.MessageConsumer) {
	logger.Infof("Starting queue consumer: %#v", messageConsumer)
	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)

	go func() {
		messageConsumer.Start()
		consumerWaitGroup.Done()
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	messageConsumer.Stop()
	consumerWaitGroup.Wait()
}

func serveEndpoints(appSystemCode string, appName string, port string, requestHandler Notifier) {

	serveMux := http.NewServeMux()

	serveMux.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	servicesRouter := mux.NewRouter()
	//todo: add new handlers here

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.Logger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	serveMux.Handle("/", monitoringRouter)

	server := &http.Server{Addr: ":" + port, Handler: serveMux}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Infof("HTTP server closing with message: %v", err)
		}
	}()

	waitForSignal()
	log.Infof("[Shutdown] notifications-push-webhooks is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Unable to stop http server: %v", err)
	}
}

func waitForSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
