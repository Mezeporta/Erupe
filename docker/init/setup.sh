#!/bin/bash
set -e
echo "INIT!"
pg_restore --username="$POSTGRES_USER" --dbname="$POSTGRES_DB" --verbose /schemas/init.sql

echo "Updating!"
if ls /schemas/update-schema/*.sql 1> /dev/null 2>&1; then
  for file in /schemas/update-schema/*.sql
  do
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f "$file"
  done
else
  echo "No update schemas found"
fi

echo "Patching!"
if ls /schemas/patch-schema/*.sql 1> /dev/null 2>&1; then
  for file in /schemas/patch-schema/*.sql
  do
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -1 -f "$file"
  done
else
  echo "No patch schemas found"
fi

echo "Database initialization complete!"
