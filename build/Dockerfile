FROM golang:alpine

WORKDIR /app
ENV GOPATH=/
COPY ./ ./

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB
ARG POSTGRES_DB_HOST
ARG POSTGRES_PORT
ENV POSTGRES_SETUP_TEST="user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB host=$POSTGRES_DB_HOST port=$POSTGRES_PORT sslmode=disable"
ENV MIGRATION_FOLDER=migrations

RUN go mod download
RUN go build -o main ./cmd/main.go
CMD goose -dir "$MIGRATION_FOLDER" postgres "$POSTGRES_SETUP_TEST" up; ./main