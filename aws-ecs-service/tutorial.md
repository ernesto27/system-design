AWS - MICROSERVICES

TODO 

- Explicar proyecto, diagrama, servicio a utilizar
- Creacion de imagenes docker,  servicio go y servicio node js
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

Para crear el servicio users,  se debe crear una carpeta llamado users y agregar una definicion de Dockerfile

$ mkdir -p docker/users && touch docker/users/Dockerfile && touch docker/users/main.go

En el archivo main.go agregar lo siguiente




