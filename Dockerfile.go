FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build args
ARG SERVICE_NAME
RUN go build -o /app/bin/service ./$SERVICE_NAME/cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/service /app/service
# Copy certs directory just in case
COPY certs /app/certs

CMD ["/app/service"]
