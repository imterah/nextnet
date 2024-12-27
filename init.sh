#!/usr/bin/env bash
if [ ! -f "backend/.env" ]; then
  cp backend/dev.env backend/.env
fi

if [ ! -d "backend/.tmp" ]; then
  mkdir backend/.tmp
fi

if [ ! -f "backend-legacy/.env" ]; then
  cp backend-legacy/dev.env backend-legacy/.env
fi

set -a
source backend-legacy/.env
source backend/.env
set +a
