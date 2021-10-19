FROM golang:alpine
COPY ./build/linux ./app
COPY ./cmd/server/config.json ./app
WORKDIR app
ENTRYPOINT ./main