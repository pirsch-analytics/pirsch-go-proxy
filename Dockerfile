FROM golang:1.20 AS build
RUN apt update && apt upgrade -y
WORKDIR /go/src/pirsch-go-proxy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" /go/src/pirsch-go-proxy/cmd/main.go && \
    mkdir /app && \
	mv main /app/server

FROM alpine
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*
COPY --from=build /app /app
WORKDIR /app

RUN addgroup -S appuser && \
    adduser -S -G appuser appuser && \
    chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
VOLUME ["/app/config.toml"]
ENTRYPOINT ["/app/server"]
