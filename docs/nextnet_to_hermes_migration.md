# NextNet to Hermes migration
## Other Environment Variables
Below are existing environment variables that need to be migrated over from NextNet to Hermes, untouched:
  * `IS_SIGNUP_ENABLED` -> `HERMES_SIGNUP_ENABLED`
  * `UNSAFE_ADMIN_SIGNUP` -> `HERMES_UNSAFE_ADMIN_SIGNUP_ENABLED`
Below are new environment variables that may need to be set up:
  * `HERMES_FORCE_DISABLE_REFRESH_TOKEN_EXPIRY`: Disables refresh token expiry for Hermes. Instead of the singular token structure used
    by NextNet, there is now a refresh token and JWT token combination.
  * `HERMES_LOG_LEVEL`: Log level for Hermes & Hermes backends to run at.
  * `HERMES_DEVELOPMENT_MODE`: Development mode for Hermes, disabling security features.
  * `HERMES_LISTENING_ADDRESS`: Address to listen on for the API server. Example: `0.0.0.0:8000`.
  * `HERMES_TRUSTED_HTTP_PROXIES`: List of trusted HTTP proxies separated by commas.
## Database-Related Environment Variables
  * `HERMES_DATABASE_BACKEND`: Can be either `sqlite` for the embedded SQLite-compliant database, or `postgresql` for PostgreSQL support.
  * `HERMES_SQLITE_FILEPATH`: Path for the SQLite database to use.
  * `HERMES_POSTGRES_DSN`: PostgreSQL DSN for Golang. An example value which should work with minimal changes for PostgreSQL databases is `postgres://username:password@localhost:5432/database_name`.
## Migration steps
1. Remove all old environment variables.
2. Add these variables:
  - `HERMES_MIGRATE_POSTGRES_DATABASE` -> `${POSTGRES_DB}`
  - `HERMES_DATABASE_BACKEND` -> `postgresql`
  - `HERMES_POSTGRES_DSN` -> `postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@nextnet-postgres:5432/${POSTGRES_DB}`
  - `DATABASE_URL` -> `postgresql://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@nextnet-postgres:5432/${POSTGRES_DB}?schema=nextnet`
  - `HERMES_JWT_SECRET` -> Random data (recommended to use `head -c 500 /dev/random | sha512sum | cut -d " " -f 1` to seed the data)
3. Switch the API docker image from `ghcr.io/imterah/nextnet:latest` to `ghcr.io/imterah/hermes-backend-migration:latest`
4. Change the exposed ports from `3000:3000` to `3000:8000`.
5. Start the Docker compose stack.
6. Go get the container logs, and make sure no errors get output to the console.
7. Copy the backup as instructed in the log file.
8. DO NOT RESTART THE CONTAINER IF SUCCESSFUL. YOU WILL LOSE ALL YOUR DATA. If the migration fails, follow the steps mentioned in the logs. You do not need to copy the DB backup if it failed to connect or read the database.
9. If successful, remove the environment variables `HERMES_MIGRATE_POSTGRES_DATABASE` and `DATABASE_URL`.
10. Switch the API docker image from `ghcr.io/imterah/hermes-backend-migration:latest` to `ghcr.io/imterah/hermes:latest`.
11. Start the backend.
## Failed Migration / Manual Restoration Steps
1. Get to step 4 in the ordinary migration setps.
2. Add the `entrypoint` option in the API compose section, and set it to `/bin/bash`
3. Add the `command` option in the API compose section, and set it to `"-c 'sleep 10000'"`
4. Get a shell in the container (likely named `nextnet-api`): `docker exec -it nextnet-api /bin/bash`
5. Copy the base64 section (excluding the `BEGIN` and `END` portions) of the backup, and run the following command to begin the transfer: `cat >> /tmp/db.json.gz.b64 << EOF`
6. Paste in the base64 data, and then press enter, type `EOF`, and then press enter again. This should return you to the shell prompt.
7. Decode the base64 backup: `cat /tmp/db.json.gz.b64 | base64 -d > /tmp/db.json.gz`
8. Run the migration script: `./entrypoint.sh`
9. When done, remove the `entrypoint` and `command` sections, and then jump to step 9 in the ordinary migration steps.
