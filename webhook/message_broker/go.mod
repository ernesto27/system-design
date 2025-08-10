module message_broker

go 1.24.0

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	queue v0.0.0
)

replace queue => ../queue
