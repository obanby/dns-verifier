FROM golang:1.18-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download && \
    cd ./cmd/verify &&\
    go build -o /usr/bin/verify

CMD ./verify

FROM alpine

COPY --from=builder /usr/bin/verify /usr/bin/verify
COPY --from=builder /app/config /app/config

ENTRYPOINT ["verify"]