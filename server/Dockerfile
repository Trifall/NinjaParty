FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o NinjaParty-Server .

FROM alpine:3.23.3

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/NinjaParty-Server .

EXPOSE 8080

ENTRYPOINT ["./NinjaParty-Server"]