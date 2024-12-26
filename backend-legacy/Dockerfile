FROM node:22.11.0-bookworm
LABEL org.opencontainers.image.source="https://github.com/greysoh/nextnet"
WORKDIR /app/
COPY src /app/src
COPY prisma /app/prisma
COPY docker-entrypoint.sh /app/
COPY tsconfig.json /app/
COPY package.json /app/
COPY package-lock.json /app/
RUN npm install --save-dev
RUN npm run build
RUN rm out/**/*.ts out/**/*.map
RUN rm -rf src
RUN npm prune --production
ENTRYPOINT sh docker-entrypoint.sh
