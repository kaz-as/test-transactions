db_host = tx-pg
db_port = 5432
db_user = root
db_pass = root
db_name = tx

conn = "host=${db_host} port=${db_port} user=${db_user} password=${db_pass} dbname=${db_name} sslmode=disable"

# Run from:

## tx-go:

migrate.up:
	goose -dir migrations postgres ${conn} up

migrate.down:
	goose -dir migrations postgres ${conn} down

migrate.status:
	goose -dir migrations postgres ${conn} status

migrate.new:
	goose -dir migrations postgres ${conn} create ${migration_name} sql

run:
	go build -o bin/tx_from_docker github.com/kaz-as/test-transactions/cmd/app && \
    ./bin/tx_from_docker

## tx-gen:

gen:
	swagger generate server -A tx --spec docs/swagger.yml --exclude-main && \
    rm restapi/configure_tx.go restapi/server.go
