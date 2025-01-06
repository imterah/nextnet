FROM golang:latest AS build
WORKDIR /build
COPY . /build
RUN bash build.sh
FROM busybox:stable-glibc AS run
WORKDIR /app
COPY --from=build /build/backends.prod.json /app/backends.json
COPY --from=build /build/api/api /app/hermes
COPY --from=build /build/sshbackend/sshbackend /app/sshbackend
ENTRYPOINT ["/app/hermes", "--backends-path", "/app/backends.json"]
