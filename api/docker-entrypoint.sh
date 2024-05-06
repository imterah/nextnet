#!/bin/bash
export NODE_ENV="production"
export DATABASE_URL="postgresql://$POSTGRES_USERNAME:$POSTGRES_PASSWORD@nextnet-postgres:5432/$POSTGRES_DB?schema=nextnet"

echo "Welcome to NextNet."
echo "Running database migrations..."
npx prisma migrate deploy
echo "Starting application..."
npm start