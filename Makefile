# Variables
user := admin
password := admin

# Commands
scylla_build_image:
	sudo rm -rf ./scylla_backup && docker build -t scylla_main:latest .

scylla_new_dangerous: 
	docker-compose down --volumes && sudo rm -rf ./scylla_backup && docker-compose up -d

scylla_start:
	docker-compose up -d

scylla_close:
	docker-compose down

scylla_cqlsh:
	docker exec -it scylla1 cqlsh -u $(user) -p $(password)

scylla_nodestatus:
	docker exec -it scylla1 nodetool status

scylla_repair:
	docker exec scylla1 nodetool repair system_auth && docker exec scylla2 nodetool repair system_auth && docker exec scylla3 nodetool repair system_auth

scylla_init:
	docker exec scylla1 cqlsh -u cassandra -p cassandra -e "CREATE ROLE admin WITH PASSWORD = 'admin' AND SUPERUSER = true AND LOGIN = true;"
	docker exec scylla1 cqlsh -u $(user) -p $(password) -e "DROP ROLE IF EXISTS cassandra"

scylla_create_tables: 
	cqlsh -u $(user) -p $(password) -f ./migrations/user.cql
	
.PHONY: scylla_build_image scylla_new_dangerous scylla_start scylla_close scylla_cqlsh scylla_nodestatus scylla_repair scylla_init

