FROM golang:1.17-alpine AS build

WORKDIR /github.com/kas2000/service-todo
COPY . $SRC_DIR
RUN go mod download
RUN GO111MODULE=on  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /service-todo

FROM alpine
WORKDIR /app
COPY --from=build /service-todo /app
COPY local.env /app

EXPOSE 8080
ENTRYPOINT ["./service-todo", "-c", "local.env"]