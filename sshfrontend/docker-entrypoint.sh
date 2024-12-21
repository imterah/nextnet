#!/bin/bash
export NODE_ENV="production"

if [[ "$SERVER_BASE_URL" == "" ]]; then
  export SERVER_BASE_URL="http://nextnet-api:3000/"
fi

npm start
