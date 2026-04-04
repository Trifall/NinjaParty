FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o NinjaParty-Server .

FROM alpine:3.23.3

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/NinjaParty-Server .

EXPOSE 8080

ENTRYPOINT ["./NinjaParty-Server"]
