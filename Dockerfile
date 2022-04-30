FROM golang:1.18 AS build
RUN apt-get update && \
    apt-get upgrade -y && \
    curl -sL https://deb.nodesource.com/setup_14.x -o nodesource_setup.sh && bash nodesource_setup.sh && \
    apt-get install -y nodejs
WORKDIR /go/src/pirsch-go-proxy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" /go/src/pirsch-go-proxy/main.go && \
    mkdir /app && \
	mv main /app/server

RUN cd /go/src/pirsch-go-proxy/js && \
    bash -c "npm i" && \
    bash -c "npm run build"
RUN mv /go/src/pirsch-go-proxy/js /app/js

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
