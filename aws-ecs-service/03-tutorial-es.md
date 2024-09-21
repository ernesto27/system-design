# PARTE 3 

En esta sección del tutorial vamos a enfocarnos en la creación de un cluster, configuraciones y los servicios (users, products) en ECS.

### Crear execution role

**Ir a IAM -> Access management -> Roles -> Crear rol**

- **Tipo de entidad de confianza:** AWS service
- **Caso de uso:** Elastic Container Service -> Siguiente
- **Nombre:** ecs-task-execution-role 

Click en crear rol

![Image](images/role-es.png)

Una vez terminado este paso ir a detalle del rol ir al detalle del rol y agregar una politica de permisos.

**Políticas de permisos -> Agregar permisos**

Agregar lo siguiente:

- AmazonECSTaskExecutionRolePolicy
- CloudWatchLogsFullAccess


### Creación de task definition

Antes de crear el cluster ECS debemos crear un "task-definition", esto lo podemos definir como un template en el cual definimos el contenedor, tipo de recursos, logs y configuraciones que van a estar asociadas a un servicio en ECS.

**Ir a ECS -> Definiciones de tareas -> Crear una nueva definición de tareas**

#### Servicio users

- **Familia de definición de tareas:**  service-users-td
- **Requisitos de infraestructura:** AWS fargate
- **Sistema operativo/arquitectura:** x86_64
- **CPU:** .25 vCPU
- **Memoria:** .5 GB
- **Rol de ejecución de tareas:** Seleccionar el rol creado anteriormente "ecs-task-execution-role"

Contenedor:

- **Nombre:** container-users
- **URI de image:**  URI de la imagen en ECR con su tag correspondiente, por ejemplo 666.dkr.ecr.us-west-2.amazonaws.com/service-users:v0.0.1
- **Contenedor escencial:** si 
- **Utilizar la recopilación de registros de CloudWatch:** habilitar opcion

Dejar las demas opciones por default y hacer click Crear.


![Image](images/task-definition-create.png)

#### Servicio products

Hacer los mismos pasos anteriores para el servicio products, pero cambiando lo siguiente.

- **Nombre:** container-products
- **URI de image:**  URI de la imagen en ECR con su tag correspondiente, por ejemplo 666.dkr.ecr.us-west-2.amazonaws.com/service-products:v0.0.1


### Crear cluster ECS 

> Un clúster ECS (Elastic Container Service) es orquestador de contenedores gestionados por AWS, que permite ejecutar y escalar aplicaciones en contenedores usando servicios como Fargate o EC2.

**Para esto debemos ir a ECS -> Clusters -> Crear cluster**

- **Nombre del clúster:** ecs-cluster
- **Infraestructura:** Fargate

Dejar las demás opciones por default y click en Crear.


### Crear servicios ECS

**Ir a detalle del cluster creado anteriormente y click en Crear**

#### Servicio users

Entorno
- **Opciones informáticas:**  Tipo de lanzamiento
- **Tipo de lanzamiento:** Fargate
- **Version de la plataforma:** Latest

Configuración de implementación

- **Tipo de aplicación:** Servicio
- **Definición de tareas:**  Seleccionar task definition "service-users-td"
- **Revision**: Mas reciente
- **Nombre del servicio:** service-users-ecs
- **Tipo de servicio:**: Replica
- **Tareas deseadas:** 1

Conexión de servicio - este servicio nos va a permitir conectar los servicios usando un DNS interno para tener baja latencia de conexion.

- **Configuración de conexión de servicio**: Cliente y servidor
- **Espacion de nombres**: ecs-tutorial
- **Agregar mapeos de puertos y aplicaciones**
	- **Alias de puerto:** Selccionar contenedor
	- **Deteccion:** users
	- **DNS:** users
	- **Puerto:** 3000

Este definicion nos va a permitir que los servicios que esten el cluster se puedan comunicar a este servicio mediante esta URL http://users:3000


Redes

- **VPC:** default
- **Subredes:** Seleccionar us-west-1a, us-west-1b
- **Grupo de seguridad:** Seleccionar los siguientes security groups
	- container-lb-sg
	- container-3000 
- **IP publica:** Activado


Balanceo de carga

- **Tipo de balanceador de carga:** Balanceador de carga de aplicaciones
- **Contenedor**: default
- **Balanceador de carga:** Usar un balanceador de carga existente, seleccionar el load balancer creado anteriormente
- **Agente de escucha:** Utilizar un agente de escucha existente, seleccionar el valor 8000:HTTP
- **Grupo de destino:** Utilizar grupo existente,  seleccionar el target group creado anteriormente "service-users-tg"


Escalado automático de servicios

- **Cantidad minima de tareas:** 1
- **Cantidad maxima de tareas:** 3
- **Tipo de politica de escalado:** Seguimiento de destino
	- **Nombre de la politica:** cpu
	- **Métrica de servicio de ECS**: EcsServiceAverageCPUUtilization
	- **Valor de destino:** 70

Con esta configuracion el servicio va a ejecutar la accion de escalamiento cuando el uso de CPU sea mayor al 70% ( scale out) y va a reducir la cantidad de tareas cuando el uso de CPU sea menor al 70% (scale in).

Click en Crear


![Image](images/ecs-service-es.png)

### Servicio products

La creacion del servicio products es similar al servicio users,  pero con las siguientes modificaciones.

Configuración de implementación
- **Nombre del servicio:** service-products-ecs
- **Definición de tareas:**  Seleccionar task definition "service-products-td"

Conexión de servicio
- **Deteccion:** products
- **DNS:** products

Balanceo de carga
- **Grupo de destino:** Utilizar grupo existente,  seleccionar el target group creado anteriormente "service-products-tg"
- **Agente de escucha:** Utilizar un agente de escucha existente, seleccionar el valor 8001:HTTP
- **Grupo de destino:** Utilizar grupo existente,  seleccionar el target group creado anteriormente "service-products-tg"

Una vez creado estos servicios deberiamos poder acceder desde el DNS de load balancer

Users: http://[YOURLOADBALANCERDND]:8000

Products: http://[YOURLOADBALANCERDND]:8001

### Actualizar servicios.

Para poder verificar que los servicios se puedan conectar entre si usando "service connect", debemos actualizar el codigo de ambos servicios

services/users/main.go
```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("service users version " + os.Getenv("API_VERSION")))
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		type User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		users := []User{
			{ID: 1, Name: "John Doe", Email: "jhon@gmail.com"},
			{ID: 2, Name: "Jane Doe", Email: "jane@gmail.com"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	r.Get("/service-products", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(os.Getenv("SERVICE_PRODUCTS") + "/products")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})

	http.ListenAndServe(":3000", r)
}
```

Ejecutamos el script deploy.sh actualizando la version de la imagen

```sh
deploy.sh <registry_name> <image_name:version> <region>
```

services/products/main.go
```go
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("service products API version " + os.Getenv("API_VERSION")))
	})

	r.Get("/products", func(w http.ResponseWriter, r *http.Request) {
		type Product struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Price int    `json:"price"`
		}

		products := []Product{
			{ID: 1, Name: "Laptop", Price: 1000},
			{ID: 2, Name: "Mouse", Price: 20},
			{ID: 3, Name: "Keyboard", Price: 50},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)

	})

	r.Get("/service-users", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(os.Getenv("SERVICE_USERS") + "/users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})
	http.ListenAndServe(":3000", r)
}
```

Para ver estos cambios reflejados tenemos que actulizar la definicion de tareas de los servicio, 
para esto ir a la seccion:

ECS -> Definiciones de tareas -> Seleccionar la definicion de tareas -> Crear una revision de tareass

En la la seccion contenedor buscar la definicion de URI y actulizar con la version correspondiente.


![Image](images/task-definition-update.png)

Hacer el mismo proceso tanto para "service-users-td" y "service-products-td"


Para aplicar el nuevo cambio en las tareas,  debemos actualizar el servicio en particular.

Ir a ECS -> Clusters -> ecs-cluster -> service-users -> Actualizar 

Seleccionar la ultima version de la definicion de tareas y click en Actualizar servicio.
Realizar el mismo proceso para el servicio products.

![Image] (images/service-update.png)


Verificar que la conexion entre servicios este funcionando correctamente


```sh
curl http://[YOURLOADBALANCERDNS]:8000/service-products
# Response 
[{"id":1,"name":"Laptop","price":1000},{"id":2,"name":"Mouse","price":20},{"id":3,"name":"Keyboard","price":50}]

curl http://[YOURLOADBALANCERDNS]:8001/service-users
# Response
[{"id":1,"name":"John Doe","email":"jhon@gmail.com"}, {"id":2,"name":"Jane Doe","email":"jane@gmail.com"}]
```


### Pruebas de stress tests, autoscaling


Actualmente los servicios estan configurados para escalar de acuerdo al uso de CPU y memoria, para poder probar esto debemos actualizar subir una nueva version de la imagen y posterior vamos a realizar pruebas de stress utilizando una herramienta llamada vegeta.


En el archivo services/users/main.go agregar lo siguiente

```go

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

r.Get("/stress-test", func(w http.ResponseWriter, r *http.Request) {
	limit := 1000000
	for i := 2; i <= limit; i++ {
		if isPrime(i) {
			fmt.Printf("%d is a prime number\n", i)
		}
	}

	w.Write([]byte("Stress test completed"))
})

```

Este codigo agrega un nuevo endpoint llamado /stress-test en el cual vamos a crear un loop que va a calcular los numeros primos hasta el numero 1000000, esta funcion va a consumir mucho CPU el cual va a disparar el autoscaling del servicio.

Hacer deploy, y actualizar la definicion de tareas y servicios en ECS.


Instalar herramienta vegeta siguiendo las instrucciones de este link.

https://github.com/tsenart/vegeta

Ejecutar el siguiente comando, el cual va a realizar un ataque al servicio users por 120 segundo y va a mostrar los resultados en la terminal.

```sh
echo "GET http:/[YOURLOADBALANCERDNS]:8000/stress-test" | vegeta attack -duration=120s | tee results.bin | vegeta report
```

Si todo sale como lo esperado,  deberiamos ver en el dashboard de ECS que la cantidad de instancias del servicio users se incremento a 2,  esto se puede ver en la seccion ECS -> Clusters -> ecs-cluster -> service-users.

![Image](images/ecs-autoscaling-es.png)

Despues de unos minutos de terminado el stress-test, el servicio va a realizar un scale in y va a volver a la cantidad de tareas original, que en nuestro caso esta definido en 1 instancia.

Se puede ver el estado de la alarma que genere el autoscaling en la seccion CloudWatch -> Alarmas -> Todas las alarmas