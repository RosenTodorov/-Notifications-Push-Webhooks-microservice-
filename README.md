# Notifications-Push-Webhooks-microservice

Description: Implement reading messages from Kafka via REST proxy - Read messages from Kafka on PostCMSPublicationEvent in separate consumer group. Also implement HTTP client that notifies clients about new publication with standard HTTP POST request and implement loading configuration about webhook addresses and authentication - create Custom Secret in K8S via Helm for this service. Inside the secret we will have JSON format. Create the project with: Helm config, Jenkins config and setup in CircleCI. 
Technologies:  Go programming, Docker, Kubernetes, Helm, Secrets, Kafka, Jenkins, CircleCI 
Development Tools: Gogland, Visual Studio Code, Vim-go plugin.
