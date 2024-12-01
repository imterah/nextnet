#!/usr/bin/env bash
owned_docker=1

# Test if Postgres is up
lsof -i:5432 2> /dev/null > /dev/null

if [ $? -ne 0 ]; then
  owned_docker=0
  docker compose -f dev-docker-compose.yml up -d
fi

if [ ! -f "api/.env" ]; then
  cp api/dev.env api/.env
fi

if [ ! -d "api/node_modules" ]; then
  pushd api > /dev/null
  npm install --save-dev
  npx prisma migrate dev
  popd > /dev/null
fi

source api/.env

on_exit() {
  cd $(git rev-parse --show-toplevel)

  if [ $owned_docker -ne 0 ]; then
    return
  fi

  docker compose -f dev-docker-compose.yml down
}

trap "on_exit" exit
