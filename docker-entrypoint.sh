#!/bin/bash
echo "Welcome to NextNet."
echo "Running database migrations..."
npx prisma migrate deploy
echo "Starting application..."
npm start