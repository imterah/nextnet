FROM golang:latest AS build
WORKDIR /build
COPY . /build
RUN cd backend; bash build.sh
FROM busybox:stable-glibc AS run
WORKDIR /app
COPY --from=build /build/backend/backends.prod.json /app/backends.json
COPY --from=build /build/backend/api/api /app/hermes
COPY --from=build /build/backend/sshbackend/sshbackend /app/sshbackend
ENTRYPOINT ["/app/hermes", "--backends-path", "/app/backends.json"]
