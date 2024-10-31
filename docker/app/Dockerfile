# Golang BE Builder
FROM golang:1.21.0-alpine AS be-builder
RUN apk update
RUN apk add curl git gcc musl-dev openssl-dev bash
WORKDIR /src
COPY ./ ./
COPY --from=fe-builder /usr/src/app/build ./web/build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./api ./cmd/api/main.go

# Prod
FROM alpine
RUN mkdir -p /docs
COPY --from=be-builder /go/src/docs /docs
COPY --from=be-builder /go/src/api /
COPY --from=be-builder /go/src/templates /
COPY ./scripts/wait-for-it.sh /
COPY ./scripts/start.sh /

RUN apk add --update bash
RUN chmod +x /start.sh /wait-for-it.sh

ENTRYPOINT [ "/start.sh" ]