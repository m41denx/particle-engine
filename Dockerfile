FROM golang:1.25 AS builder
LABEL authors="M41den"

RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w -X main.BuildTag=$(git rev-parse --short HEAD) -X main.BuildDate=$(date '+%Y-%m-%dT%H:%M') -X main.Version=${VERSION}" \
     -o particle "./cmd"

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Europe/Moscow
COPY --from=builder /app/particle /app/particle

CMD /app/particle