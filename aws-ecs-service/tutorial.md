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


# Crear ECR repositorio

La region que vamos a utilizar es Viriginia,  es importante que todos los servicios que requieran una seleccion de region esten creados en el mismo lugar.

Para poder subir nuestras imagenes a AWS necesitamos un repositorio en ECR, para esto debemos ir a la consola de AWS y buscar por ECR.
Ir a seccion -> Private registry -> Repositorios -> Crear repositorio

Como nombre poner "service-users",  modificar mutablidad a "Inmutable", esto lo que va a generar es que no se puedan subir imagenes con el mismo tag, para evitar problemas de diferentes versiones del servicio con el mismo tag,  obliga a que cada nuevo deploy este asociado a un tag diferente.

Dejamos las demas opciones en default y click en Crear repositorio.

TODO PONER IMAGEN

#### Instalacion CLI aws
Necesitamos tener instalado en nuestra maquina la herramienta CLI de aws,  sigue las instrucciones de acuerdo a tu sistema operativo en el siguient link.

https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html 

Una vez instalado la CLI de aws,  debemos configurar las credenciales de acceso que creamos anteriormente para esto ejecutamos el siguiente comando en una terminal.

```sh
aws configure
```
En la terminal debemos utilizar las credenciales (access key, private key) creadas en el paso anterior y como region default seleccionar "us-east-1" (Virginia).

#### Uplaod imagen a ECR

El la carpeta docker/users agregar un archivo llamado deploy.sh con el sigueinte contenido.

deploy.sh
```sh
#!/bin/bash

if [ $# -ne 3 ]; then
    echo "Usage: deploy.sh <registry_name> <image_name> <region>"
    exit 1
fi

REGISTRY_URI=$1
REGISTRY_NAME=$2
REGION=$3

aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $REGISTRY_URI

docker build -t $IMAGE_NAME .

docker tag $IMAGE_NAME $REGISTRY_URI/$REGISTRY_NAME

docker push $REGISTRY_URI/$REGISTRY_NAME
```

Este script recibe tres argumentos.
URI de repositorio
Nombre del repositorio
Region de aws

Antes de ejecutar el script debemos darle permisos de ejecucion.

```sh
chmod +x deploy.sh
```

Para subir la imagen a ECR ejecutamos el siguiente comando, modificando los valores correspondientes generandos en su cuenta de AWS, el valor URI del repositorio se encuentra en el dashboard de ECR.

```sh
./deploy.sh 666.dkr.ecr.us-west-2.amazonaws.com service-user:v0.0.1 us-west-1
```

Si todo salio correctamente deberiamos ver la version de la imagen en el dashboard de ECR.


## Servicio products

Debido a que el contenido del servicio productos es practicamente igual al de users ( solo va a cambiar el contenido del archivo main.go) vamos a ejecutar un comando para hacer una copia de users y modificar los archivos necesarios.

```sh
cp -ra services/users services/products
```

Modificar la respuesta del endpoint en el archivo services/procuts/main.go

```go
///////////////////////////////////////////////////////////////
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("service products API version " + os.Getenv("API_VERSION")))
})
///////////////////////////////////////////////////////////////
```

Crear un nuevo repositorio privado en ECR con el nombre service-products y realizar un deploy con el script deploy.sh

```sh
./deploy.sh <registry-uri> <registry-name> <region>  
```



# PARTE 2 

## Creacion de load balancer

En este proyecto vamos a utilizar un load balancer para distribuir el trafico entre los servicios users y products que postierormente vamos a crear en ECS, pero antes de eso debemos configurar otros servicios requeridos.


#### Crear security groups

El primer security group que vamos a crear va a estar asociado al load balancer y es el que va a permitir el trafico de entrada tanto al puerto 8000 como al 8001.

Para esto tenemos que ir a  EC2 -> Red y seguridad -> Security groups -> Crear security group

Como nombre ponemos el valor "load-balancer-sg" y una descripcion "security group para accesso externo al load balancer" , dejamos el valor de VPC por default

Como regla de entrada agregamos estas dos configuraciones.

- Tipo: TCP personalizado 
- Protocolo: TCP
- Intervalo de puertos: 8000
- Origen:  Anywhere - Esto va a permitir el trafico de cualquier IP al puerto 8000

--- 
---
---

- Tipo: TCP personalizado 
- Protocolo: TCP
- Intervalo de puertos: 8001
- Origen:  Anywhere - Esto va a permitir el trafico de cualquier IP al puerto 8000


En reglas de salida dejamos el valor default, el cual permite el trafico de salida a cualquier origen y hacemos click en Crear security group.

El segundo security group que vamos a crear va a estar asociado a los servicios de ECS, este debe estar asociado al segundo security group creado anteriormente,  para que todo el trafico de entrada provenga desde el load balancer.

Como nombre ponemos el valor "container-sg" y una descripcion "security group para la comunicacion entre el load balancer y los contenedores" , dejamos el valor de VPC por default

- Tipo: Todos los TCP
- Protocolo: TCP
- Intervalo de puertos: 0-65535
- Origen:  Personalizada
- Origen valor: Buscar el nombre del security group creado anteriormente "load-balancer-sg" y seleccionarlo.

Dejar las reglas de salida por default y click en Crear security group.


#### Crear target group

Ir a seccion EC2 -> Equilibrio de carga -> Grupos de destino -> Crear grupo de destino

Elegir un tipo de destino: Direcciones IP
Nombre del grupo de destino: service-users-tg
Protocolo: HTTP - 8000
Tipo de direccion IP: IPv4
VPC: default
Version del protocolo: HTTP1

Dejar los demas valores por defaul,  click en siguiente.

En seccion especificar direcciones IP y definir puertos,  click en eliminar IP definida por default,  estos valores
se van a crear dinamicamente cuando creemos el servicio en ECS.


Hacer el mismo paso anterior para el servicio products,  pero cambiando el nombre del grupo de destino a "service-products-tg" y el puerto a 8001.


#### Crear load balancer

Ir a seccion EC2 -> Equilibrio de carga -> Balanceadores de carga -> Crear balanceador de carga

Seleccionamos el tipo  "Balanceador de carga de aplicaciones" el cual nos va a permitir balancear el trafico hacia los servicios en ECS y configurar reglas de enrutamiento mas avanzadas.

Click en crear y definir estos valores.

- **Nombre del balanceador de carga:**  load-balancer-ecs
- Esquema: Expuesto a Internet 
- Tipo de dirección IP del equilibrador de carga: IPv4
- VPC: default
- Mapeos: us-west-1a, us-west-1b
- Grupos de seguridad: seleccionar el security group creado anteriormente "load-balancer-sg"
- Agentes de escucha y redireccionamiento: 
	- Protocolor: HTTP
	- Puerto: 8080
	- Accion predeterminada: service-user-sg
- Nuevo agentes de escucha
	- Protocolor: HTTP
	- Puerto: 8081
	- Accion predeterminada: service-products-sg


La creacion del load balancer tarda unos minutos , una vez terminado el proceso, podemos acceder al DNS del load balancer generado por AWS.

Si ingresamos a esta URL en el puerto 8000,  deberiamos ver un mensaje de 503 Service Temporarily Unavailable, nuestro proximo paso es configurar el cluster ECS para hacer uso de este servicio.


# PARTE 3 
## Crear cluster ECS















#### 











