# get base golang images
FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /app

COPY . /app
COPY ./templates /templates

WORKDIR /app

RUN CGO_ENABLED=0 go build -o MailerApp ./cmd/api

RUN chmod +x /app/MailerApp

#build a tiny docker image

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/MailerApp /app
COPY --from=builder ./templates /templates

CMD ["/app/MailerApp"]