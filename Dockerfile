# syntax=docker/dockerfile:1

FROM golang

# Set destination for COPY
WORKDIR /app

# install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY config.yml ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /executable/worker cmd/main/app.go

# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can (optionally) document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
#EXPOSE 8080

# Run
WORKDIR /executable
CMD ["/executable/worker"]