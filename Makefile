.PHONY: test deps docker release

test:
	go test -cover -race github.com/pirsch-analytics/pirsch-go-proxy/pkg/proxy

deps:
	go get -u -t ./...
	go mod tidy
	go mod vendor

docker: test
	docker build -t pirsch/proxy:$(VERSION) -f build/Dockerfile .
	docker push pirsch/proxy:$(VERSION)

release: test
	mkdir -p pirsch
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" cmd/main.go
	CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" cmd/main.go
	mv main pirsch/pirschproxy
	mv main.exe pirsch/pirschproxy.exe
	cp config/config.toml pirsch
	cp README.md pirsch
	cp CHANGELOG.md pirsch
	cp LICENSE pirsch
	zip -r "pirsch_proxy_v$(VERSION).zip" pirsch
	rm -r pirsch
