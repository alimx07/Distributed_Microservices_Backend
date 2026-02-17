#!/usr/bin/env bash
set -e

PRIMARY_HOST="postgres_primary"
PRIMARY_PORT=5432
REPL_USER="physical_rep"
PGDATA=${PGDATA:-/var/lib/postgresql/data}


echo "Waiting for primary ${PRIMARY_HOST}:${PRIMARY_PORT}..."
until pg_isready -h "$PRIMARY_HOST" -p "$PRIMARY_PORT" -U postgres; do
  sleep 1
done
echo "Primary is ready."


if [ -s "$PGDATA/PG_VERSION" ]; then
  echo "Existing PGDATA found; removing to create base backup..."
  rm -rf "${PGDATA:?}/"*
fi


echo "Running pg_basebackup..."
export PGPASSWORD="${PGPASSWORD:-replicator_pass}"
pg_basebackup -h "$PRIMARY_HOST" -p $PRIMARY_PORT -D "$PGDATA" -U "$REPL_USER" -v -P -X stream -R

echo "Base backup complete. primary_conninfo and standby config written."


exec docker-entrypoint.sh postgres
