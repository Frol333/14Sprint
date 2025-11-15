FROM golang:1.24.2-alpine AS builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -o app ./…
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /build/app /app/server
COPY --from=builder /build/web /app/web
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/todo.db
ENV TODO_PASSWORD=""
EXPOSE 7540
CMD ["./server"]