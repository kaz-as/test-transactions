# TEST - Transactions

## Run

To read documentation and try requests, run `make run` from `zip-go` container. Docs will be at http://localhost:8080/docs

## Generate code from documentation

Run `make gen` from `tx-gen` container.

Do not create files `restapi/configure_tx.go`, `restapi/server.go` by your own: they are deleted by `make gen`.

## Migrations

Should be run from `tx-go` container.

Up:
```bash
make migrate.up
```

Down:
```bash
make migrate.down
```

Status:
```bash
make migrate.status
```

Create new migration with name `new123`:
```bash
make migrate.new migration_name=new123
```

## TODO
* Implement handlers themselves
* Add specific errors to swagger
* Add authentication and authorization
* Create indices to SQL tables
* Add tests
