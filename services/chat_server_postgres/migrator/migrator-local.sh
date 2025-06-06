#!/bin/bash
source local.env

#export MIGRATION_DIR="./migrations"
#export MIGRATION_DSN="host=chat-server_pg port=5432 dbname=$POSTGRES_DB user=$POSGRES_USER password=$POSTGRES_PASSWORD sslmode=disable"

sleep 2 && goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v