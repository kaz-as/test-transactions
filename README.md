# TEST - Transactions

## Overview

A simple transaction processing. A transaction is an action of transferring money from one account to another.

Firstly, the processing owns `1 000 000 000`.

When creating a user, he can obtain some money, transferred from the processing account.

A transaction cannot proceed (logical OR):
* sender = receiver;
* any of sender or receiver does not exist;
* the sender does not have enough money;
* the amount of money being transferred is negative;
* when the receiver obtains transaction he would have too much money (more than `9 223 372 036 854 775 807`).

Transaction on different accounts are running in parallel. Each has its own queue: there are no conflicting DB locks
between DB-transactions that have no common accounts.

## TODO
* **Add tests**
* Add specific errors to swagger
* Add authentication and authorization
* Create indices to SQL tables

## Run
To read swagger documentation and try requests, run `make run` from `tx-go` container. Docs will be at http://localhost:8081/docs

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
