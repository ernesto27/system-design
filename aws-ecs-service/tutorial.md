AWS - MICROSERVICES

TODO 

- Explicar proyecto, diagrama, servicio a utilizar
- Creación de imagenes docker,  servicio go users, products
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
- Servicio task job cron
- Crear RDS servicio mysql, connect and add mysql tablas para testing
- Actualizar docker images , connect to mysql, deploy ECR
- Actualizar servicios
- Pruebas de stress tests,  probar autoscaling
- configuracion dominio, https, load balancer setup
- WAF firewall
- Eliminacion de servicios.
- CLI AWS deploy servicios
- CD - CI pipeline











# PARTE 4

- Configuracion dominio,  https, load balancer setup
- Agregar WAF firewall

Al momento tenemos creado un load balancer al cual podemos acceder via un DNS usando http de esta manera.

http://[YOURLOADBALANCERDNS]:8000


En esta parte del tutorial el objetivo es asociar un nombre de dominio a nuestro load balancer y configurar un certificado SSL para poder acceder a los servicios de manera segura.

En mi caso voy a usar un dominio creado en NameCheap,  pero pueden usar cualquier otro proveedor de dominios que tengan en el cual puedan agregar registro DNS.

### Crear certfiicado SSL 

Para poder crear un certificado SSL en AWS necesitamos ir a la seccion de ACM (Amazon Certificate Manager) y click en Solicitar.

En nombres de dominio vamos a agregar dos valores

- yourdomain.com
- *.yourdomain.com

El * nos permite poder validar cualquier subdominio que se cree en el dominio principal.
Por ejemplo: api.yourdomain.com, users.yourdomain.com, etc.

![Image](images/acm-create.png)

Dejar las demas opciones por default (Validación de DNS) y click en solicitar.

Ir a detalle de certificado y copiar los siguintes valores.

Nombre CNAME y valor CNAME 

Ir a la configuracion DNS del registro y agregar un nuevo registro CNAME con el nombre y valor copiado anteriormente.

- Host: _AAAA1111
- Value: _AAAA1111.acm-validations.aws.

En el caso de usar NameCheap tener en cuenta que el valor del host copiado desde ACM no debe tener el dominio.

Por ejemplo.

_AAAA1111.mydomain.com.

Debe quedar de la siguiente manera

_AAAA1111

> Nota: Tener en cuenta que en otros proveedores de DNS si puede ser necesario agregar al registro CNAME el mismo valor que nos da ACM.


Este proceso de validacion via DNS puede tardar unos 5 minutos aproximadamente,  en la seccion de ACM se va a ver el estado del certificado como "Emitido".


### Configurar dominio

En este tutorial vamos a utilizar el servicio de DNS namecheap, de todas maneras el proceso es muy similar en otros proveedores.

Ir a detalle de dominio - Advanced DNS 

Debemos agregar dos registro de CNAME con estos valores

**Host:** @ 

**Value:** [YOURLOADBALANCERDNS]

--- 

**Host:** www

**Value:** [YOURLOADBALANCERDNS]

![Image](images/dns-cname.png)


Una vez realizado estos pasos podemos acceder a nuestro servicio usando nuestro dominio, tener en cuenta que el proceso de propagacion de DNS puede tardar unos minutos.

http://[YOURDOMAIN]:8000


### Configuracion load balancer SSL

Previo a la configuracion del load balancer,  debemos modificar el security group del load balancer para permitir el trafico HTTPS.

Ir a EC2 -> Red y seguridad -> Security groups -> click en load-balancer-sg -> Editar reglas de entrada

Agregar regla 

- Tipo: HTTPS
- Protocolo: TCP
- Intervalo de puertos: 443
- Origen:  Anywhere - 0.0.0.0/0

Click en Guardar reglas

![Image](images/sg-https.png)


Para poder utilizar HTTPS en nuestro dominio debemos agregar una configuracion en nuestro load balancer.  
para esto debemos ir a la seccion de EC2 -> Load Balancers -> load-balancer-ecs -> Agente de escucha y reglas -> Agregar agente de escucha

Configuracion agente de escucha:
- Protocolo: HTTPS
- Puerto: 443

Acciones Predeterminadas

- Renviar a grupos de destino
- Grupo de destino: service-users-tg


![Image](images/lb-ssl.png)


**Configuracion de agente de escucha seguro**

Certificado de servidor SSL/TLS predeterminado

- Origen del certificado: de ACM
- Certificado de ACM: Seleccionar certificado creado anteriormente en ACM

Dejar los demas valores por default y click en Agregar

Despues de terminada esta configuracon podemos ingresar a nuestro dominio de esta manera

https://[YOURDOMAIN]

### Configuracion subdominios servicios

Para poder acceder a los servicios creados en ECS,  vamos a configurar subdominios para cada servicio.

Por ejemplo:

users: https://users.[YOURDOMAIN]

products: https://products.[YOURDOMAIN]

Como primer paso debemos ir a la configuracion de DNS de nuestro dominio y agregar dos registros CNAME con los siguientes valores.

- Host: users
- Value: [YOURLOADBALANCERDNS]

---

- Host: products
- Value: [YOURLOADBALANCERDNS]

Posteriormente debemos modificar la configuracion del load balancer en AWS

Ir a la seccion de EC2 -> Load Balancers -> load-balancer-ecs -> Agente de escucha y reglas

Seleccionar el agente de escucha HTTPS -> Administrar reglas -> Agregar una regla

![Image](images/lb-rules.png)

Nombre y etiquetas  

- Nombre: users-rules

Agregar condicion

Encabezado de Host

- users.[YOURDOMAIN]	

Tipos de accion 

Reenviar a grupos de destino

- Prioridad 1
- Grupos de destino: service-users-tg



Crear los mismo pasos para el servicio products, modificando el valor del host en la condicion a products.[YOURDOMAIN] y el grupo de destino a service-products-tg.


### WAF firewall

Si bien tenemos nuestro servicio y load balancer funcionando, es importante tener en cuenta la seguridad de los servicios que estamos exponiendo al publico, para esto vamos a configurar un WAF (Web Application Firewall) el cual nos va a permitir proteger nuestros servicios de ataques comunes como ser SQL injection, XSS, etc.

En la seccion de load balancer de AWS ir Integraciones -> AWS Web Application Firewall

Click en Asociar WAF

Dejar las opciones por default, el cual utiliza 3 reglas de proteccion predefinidas,  el comportamiento por default va a ser bloquear el trafico que no cumpla con estas reglas,  click en Asociar.

![Image](images/waf.png)

# PARTE 5
En esta seccion vamos a configurar nuestros servicios para que se puedan conectar a una base de datos mysql,  para esto vamos a utilizar un servicio de AWS llamado RDS 

Este es un servicio que proporciona una base de datos relacional completamente administrada en la nube. Permite a los desarrolladores configurar, operar y escalar fácilmente una base de datos relacional en la nube sin tener que preocuparse por la infraestructura. RDS admite varios motores de bases de datos, como MySQL, PostgreSQL, Oracle, SQL Server y Amazon Aurora, backups automatizados,  configuracion en multiples zonas de disponibilidad, etc.

Para este tutorial vamos a utilizar Mysql.

Previo a la creacion de la base de datos debemos agregar un security group que permita la conexion la base de datos desde otros servicios.

### Grupo de seguridad
Ir a EC2 -> Red y seguridad -> Security groups -> Crear security group

Nombre: rds-sg
Descripcion: security group para la conexion a la base de datos
VPC: default

Reglas de entrada -> Agregar regla

- Tipo: MySQL/Aurora
- Origen: Anywhere IPV4

Click en Crear security group


### RDS
Ir al dashboard de AWS -> RDS -> Bases de datos -> Crear base de datos

![Image](images/rds-es.png)


Elegir un método de creación de base de datos: Creacion estandar

Opciones del motor
- Tipo de motor: mysql
- Version del motor: mysql 8.0.55

Plantillas: Desarrollo y pruebas

Configuracion:
- Identificador de la base de datos: ecs-tutorial
- Nombre de usuario: admin
- Administracion de credenciales: Autoadministrado
- Contraseña: [YOURPASSWORD]
- Confirmar contraseña: [YOURPASSWORD]

Configuracion de la instancia

- Unchecked: Mostrar las clases de instancia que admiten las escrituras optimizadas de Amazon RDS
- Check en incluir clases de generacion anterior ( esto nos va a permitir seleccionar una clase de instancia con menos recursos)
- Clase de instancia: db.t4.micro

Alamacenamiento: Dejar valores default

Conectividad:
- Recurso de computacion: No se conecta a un grupo de EC2
- Tipo de red: IPv4
- Acceso publico: Si
- Grupo de seguridad firewall: Elegir existente, seleccionar el security group creado anteriormente "rds-sg"

Autenticación de bases de datos: Auteenticacion de contraseña

Dejar las demas configuraciones por default ,  click en crear base de datos.



TODO IMAGEN


Conclusiones:

En este tutorial hemos creado un proyecto en el cual hemos utilizado varios servicios de AWS para un proyecto de microservicios, hemos creado dos servicios users y products, los cuales se comunican entre si y se exponen al publico a traves de un load balancer,  hemos configurado un dominio y certificado SSL para acceder a los servicios de manera segura,  ademas hemos configurado un WAF para proteger nuestros servicios de ataques comunes.

Finalmente y muy importante en caso de que no vayan a utilizar los servicios creados en este tutorial,  es elminar todos los recursos creados en AWS para evitar costos no deseados.

Espero que este tutorial les haya sido de utilidad,  cualquier duda o consulta pueden contactarme a traves de mi correo personal.
































#### 











