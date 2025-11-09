#!/bin/bash
set -e
echo "INIT!"
pg_restore --username="$POSTGRES_USER" --dbname="$POSTGRES_DB" --verbose /schemas/init.sql



echo "Updating!"

for file in /schemas/update-schema/*
do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f $file
done



echo "Patching!"

for file in /schemas/patch-schema/*
do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f $file
done
