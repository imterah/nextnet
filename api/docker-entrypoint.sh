#!/bin/bash
export NODE_ENV="production"

if [[ "$DATABASE_URL" == "" ]]; then
  export DATABASE_URL="postgresql://$POSTGRES_USERNAME:$POSTGRES_PASSWORD@nextnet-postgres:5432/$POSTGRES_DB?schema=nextnet"
fi

echo "Welcome to NextNet."
echo "Running database migrations..."
npx prisma migrate deploy
echo "Starting application..."
npm start
