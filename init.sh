if [ ! -d ".tmp" ]; then
  echo "Hello and welcome to the NextNet project! Please wait while I initialize things for you..."
  cp dev.env .env
  mkdir .tmp
fi

lsof -i:5432 | grep postgres 2> /dev/null > /dev/null
IS_PG_RUNNING=$?

if [ ! -f ".tmp/ispginit" ]; then
  if [[ "$IS_PG_RUNNING" == 0 ]]; then
    kill -9 $(lsof -t -i:5432) > /dev/null 2> /dev/null
  fi
  
  echo " - Database not initialized! Initializing database..."
  mkdir .tmp/pglock

  initdb -D .tmp/db
  pg_ctl -D .tmp/db -l .tmp/logfile -o "--unix_socket_directories='$PWD/.tmp/pglock/'" start
  createdb -h localhost -p 5432 nextnet 

  psql -h localhost -p 5432 nextnet -c "CREATE ROLE nextnet WITH LOGIN SUPERUSER PASSWORD 'nextnet';"

  npm install --save-dev
  npx prisma migrate dev

  touch .tmp/ispginit
elif [[ "$IS_PG_RUNNING" == 1 ]]; then
  pg_ctl -D .tmp/db -l .tmp/logfile -o "--unix_socket_directories='$PWD/.tmp/pglock/'" start
fi

source .env # Make sure we actually load correctly