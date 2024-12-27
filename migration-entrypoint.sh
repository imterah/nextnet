#!/usr/bin/env bash
echo "Welcome to the Hermes migration assistant."

if [ ! -f "/tmp/db.json.gz" ]; then
  echo "Exporting database contents..."
  cd /app/legacy
  node out/tools/exportDBContents.js /tmp/db.json.gz
  BACKUP_EXIT_CODE=$?

  if [ $BACKUP_EXIT_CODE -ne 0 ]; then
    echo "Failed to export database contents!"
    exit 1
  fi

  echo "!! IMPORTANT !!"
  echo "Database backup contents below:"
  echo "==== BEGIN BACKUP ===="
  cat /tmp/db.json.gz | base64
  echo "==== END BACKUP ===="
  echo "When copying, do NOT copy the BEGIN and END sections."
fi

echo "Wiping old database..."
cat >> /tmp/wipe.sql << EOF
CREATE DATABASE temp;
\c temp
DROP DATABASE $HERMES_MIGRATE_POSTGRES_DATABASE;
CREATE DATABASE $HERMES_MIGRATE_POSTGRES_DATABASE;
\c nextnet
DROP DATABASE temp;
EOF
psql "$HERMES_POSTGRES_DATABASE" < /tmp/wipe.sql
rm -rf /tmp/wipe.sql
echo "Restoring backup..."

cd /app/modern
./hermes -b ./backends.json import --bp /tmp/db.json.gz

echo "Restored backup. If this restore fails after the database has wiped, get a shell into the container,"
echo "copy the backup contents into the container (base64 decoded) at '/tmp/db.json.gz',"
echo "and rerun /app/entrypoint.sh."
echo ""
echo "If further issues continue, open an issue at 'https://git.terah.dev/imterah/hermes'."
echo "If the migration succeeded, congratulations!"

sleep 10000
