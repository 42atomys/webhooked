FROM golang:1.20-alpine AS build

WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o webhooked

FROM alpine

LABEL maintener "42Atomys <contact@atomys.fr>"
LABEL repository "https://github.com/42Atomys/webhooked"

COPY --from=build /build/webhooked /webhooked

CMD ["/webhooked", "serve"]
