module consumer_service

go 1.24.0

require (
	database v0.0.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	queue v0.0.0
)

require (
	github.com/lib/pq v1.10.9 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.24.3 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
)

replace queue => ../queue

replace database => ../database
