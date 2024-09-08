AWS - MICROSERVICES

TODO 

- Explicar proyecto, diagrama, servicio a utilizar
- Creacion de imagenes docker,  servicio go users, products
- Crear usuario IAM con permisos
- Creacion de ECR
- Instalar CLI aws , configuraion access key
- Deployar version v0.0.1 a ECR 
- Crear security groups para servicios puerto 8000 y 8001
- Crear target group para los 2 servicios
- Crear load balanacer, configuracion listeners, etc.
- Crear task definition ECS container go y nodejs 
- Crear cluster ECS 
- Crear security groups para ECS 
- Crear servicios ECS , asociarlo a load balancer
- Crear RDS servicio mysql, connect and add mysql tablas para testing
- Actualizar docker images , connect to mysql, deploy ECR
- Actualizar servicios
- Pruebas de stress tests,  probar autoscaling
- configuracion dominio, https, load balancer setup
- WAF firewall
- Eliminacion de servicios.


En este tutorial vamos a crear un proyecto en el cual vamos a utilizar microservicios,  los cuales van a estar ejecutandose en AWS.
Vamos a estar utilizando los siguentes servicios de AWS.

ECS 
Load Balancer
RDS
ERC

Los servicios se van a conectar a una base de datos mysql y ademas se van a conectar entre si intercambiando mensajes.

TODO DIAGRAMA

Docker image -  service users
Docker image -  service productus

Vamos a crear dos servicios para el proyecto,  en general el numero de este va a ser mayor, 
pero en pos de mantener simple el tutorial vamos a limitarnos al momento a solo dos servicios.

## Crear micro-servicio users container


Para crear el servicio users,  se debe crear una carpeta llamado users y agregar una definicion de Dockerfile

$ mkdir -p docker/users

En la carpeta docker/users agregar los siguientes archivos

go.mod

```
module service-users

go 1.22.3

require (
	github.com/go-chi/chi/v5 v5.1.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)
```

En este archivo definimos la version de go que vamos a utilizar para compilar el servicio y las dependecias a utilizar
en este caso serian las siguiente:
go-chi: es un router http el cual nos facilita la creacion de endpoints, middlewares, etc.
godotenv: es una libreria que nos permite definir variables de entorno en un archivo .env.

main.go

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("service users API version " + os.Getenv("API_VERSION")))
	})
	http.ListenAndServe(":3000", r)
}
```
En la funcion main hacemos una llamada a godotenv.Load() para cargar las variables de entorno definidas en el archivo .env
en caso de encontrar un error al cargar el archivo se termina la ejecucion del programa.

Despues iniciamos un router de go-chi y definimos un endpoint en la raiz del servidor el cual va a mostrar la version de la API correspondiente al valor definido en el archivo .env.


.env

```
API_VERSION=1.0.0
```

En este archivo al momento solo definimos la version de la API,  mas adelante vamos a agregar mas valores como serian los 
valores de conexion a una base de datos.

Dockerfile

```
# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /service

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /service /service
COPY .env ./

EXPOSE 3000


ENTRYPOINT ["/service"]
```

En este archivo Dockerfile vamos utilizar un concepto llamado multi-stage build,  el cual nos permite utilizar una imagen base para compilar el servicio ( golang 1.22) copiar los archivos necesarios, descargar dependencias y finalmete compilar el servicio en un binario llamado service.
La segunda imagen (gcr.io/distroless/base-debian11) es la que se va a utilizar en ECS para ejecutar la API,  esta imagen es una imagen minimalista que solo contiene lo necesario para ejecutar el binario, el cual tenemos que copiar desde el output obtenido del build previo, ademas necesitamos copiar el .env, exponenmos el puerto 3000 e iniciamos el binario ./service.


Para probar esto debemos ejecutar los siguentes comandos en una terminal
```sh
docker build  -f Dockerfile-go -t service-users:v0.0.1 .
docker run -p 3000:3000 service-users:v0.0.1
```


## Crear usuario IAM credenciales

Para poder subir nuestra imagen a ECR necesitamos un usuario en AWS con los permisos correspondientes, para esto debemos realizar los siguiente.

Ir al dashboard de nuestra cuenta de AWS e ir a la siguiente seccion.

IAM -> Administracion de usuarios -> Usuarios -> Agregar usuario

Como nombre poner lo siguiente "ecs-lb-tutorial" click en siguiente.
En seccion Establecer permisos -> Opciones de permisos seleccionar Adjuntar politicas directamente -> Crear politica
En politicas de permisos buscar por "AmazonEC2ContainerRegistryFullAccess" seleccinar la politica y click en siguiente y 
finalizar la creacion del usuario.

### Crear access key

Necesitamos un access and private key para poder comunicarnos a los servicios de AWS usando la cli, para esto debemos ir al detalle del usuario creado anteriormente,  Administracion de usuarios -> Usuarios.

Ir a tab Credenciales de seguridad -> Crear clave de acceso 
En casos de usos seleccionar Otros, click en Siguiente -> Crear clave de acceso
Copiar en un lugar seguro la clave de acceso y el secret key, ya que no se va a poder ver nuevamente.












