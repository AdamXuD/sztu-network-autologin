FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

ENV GOOS linux

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -ldflags="-s -w" -o /app/main /app/main.go

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata

ENV TZ Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/main /app/main

COPY --from=builder /app/.env /app/.env

CMD ["./main"]