#!/bin/bash
export MIGRATION_DIR="./migrations"
export MIGRATION_DSN="host=auth_pg port=5432 dbname=$POSTGRES_DB user=$POSTGRES_USER password=$POSTGRES_PASSWORD sslmode=disable"

sleep 2 && goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v