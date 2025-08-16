module consumer_service

go 1.24.0

require (
	github.com/joho/godotenv v1.5.1
	queue v0.0.0
)

require github.com/rabbitmq/amqp091-go v1.10.0 // indirect

replace queue => ../queue

replace database => ../database
