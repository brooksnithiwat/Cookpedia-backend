### Multi-stage Dockerfile for Cookpedia Backend
FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/cookpedia_backend ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/cookpedia_backend ./cookpedia_backend
EXPOSE 8080แ
ENV GIN_MODE=release
CMD ["/app/cookpedia_backend"]
