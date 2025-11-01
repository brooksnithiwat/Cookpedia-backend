### Multi-stage Dockerfile for Cookpedia Backend
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go version && go env && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -x -o /app/cookpedia_backend ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/cookpedia_backend ./cookpedia_backend
EXPOSE 8080
ENV GIN_MODE=release
CMD ["/app/cookpedia_backend"]
