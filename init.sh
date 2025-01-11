#!/usr/bin/env bash
if [ ! -d "$HOME/go/bin" ]; then
  mkdir ~/go/bin
fi

export PATH="$HOME/go/bin:$PATH"

if [ ! -f "backend/.env" ]; then
  cp backend/dev.env backend/.env
fi

if [ ! -d "backend/.tmp" ]; then
  mkdir backend/.tmp
fi

if [ ! -f "frontend/.env" ]; then
  cp frontend/dev.env frontend/.env
fi

set -a
source backend/.env
source frontend/.env
set +a
