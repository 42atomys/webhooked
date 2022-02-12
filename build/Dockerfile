FROM golang:1.17-alpine AS build

WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o dns-updater

FROM alpine

LABEL maintener "42Stellar <contact@42stellar.org>"
LABEL repository "https://github.com/42Stellar/webhooks"

COPY --from=build /build/webhooks /webhooks

CMD ["/webhooks"]
