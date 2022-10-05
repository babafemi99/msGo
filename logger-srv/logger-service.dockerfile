# get base golang images
FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o loggerApp ./cmd/api

RUN chmod +x /app/loggerApp

#build a tiny docker image

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/loggerApp /app

CMD ["/app/loggerApp"]