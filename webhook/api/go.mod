module webhook

go 1.24.0

require (
	github.com/labstack/echo/v4 v4.13.4
	github.com/stretchr/testify v1.10.0
	queue v0.0.0
)

replace queue => ../queue

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
