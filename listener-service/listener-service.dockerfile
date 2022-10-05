# get base golang images
FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o listenerApp .

RUN chmod +x /app/listenerApp

#build a tiny docker image

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/listenerApp /app

CMD ["/app/listenerApp"]