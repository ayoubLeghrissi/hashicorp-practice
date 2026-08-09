#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE product_db;
	CREATE DATABASE inventory_db;
	CREATE DATABASE cart_db;
EOSQL
