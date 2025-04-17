
## Database Setup (Docker)

To run a PostgreSQL database locally using Docker, use the following command:

```bash
docker run --name my-postgres-db -e POSTGRES_PASSWORD=yourpassword -p 5432:5432 -d postgres:16
```



### Migraciones

#### Instalar herramienta goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

#### Crear archivo de migration 

```bash
goose create --dir="migrations" newtable sql
```