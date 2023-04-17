# clique_server
Server for clique chat app built on websockets.

Install `cqlsh`
```bash
pip install cqlsh
```
## Initialize DB
1)  Create scylla docker image
```bash
make scylla_build_image
```

2) Set request capacity and run the db 
```bash
make set_request_capacity # run this command only the first time running the db or after rebooting the system
make scylla_new_dangerous
```
wait for the three nodes to be set up properly, until the status of all the three racks in UN.

```bash
make scylla_nodestatus
```

3) Create tables 
```bash
make scylla_init
make scylla_create_tables
```

4) Check if db works correctly
```bash
make scylla_cqlsh
```

## Run the API

```bash
go get -u ./...
go mod vendor
go mod tidy
go run cmd/main.go
```

## Endpoints

`localhost:5050/auth/signup`
`localhost:5050/auth/login`
`localhost:5050/auth/logout`
