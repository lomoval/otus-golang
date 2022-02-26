#!/bin/bash

DBSTRING="host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USERNAME password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB sslmode=disable"
echo "$DBSTRING"

./wait && goose postgres "$DBSTRING" up