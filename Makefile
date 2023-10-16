.PHONY: build
build:
	go build -o mackerel-plugin-mongodb

.PHONY: test
test: testgo testmetric

.PHONY: testmetric
metrictest: build
	go install github.com/lufia/graphitemetrictest/cmd/graphite-metric-test@latest
	./test.sh

.PHONY: testgo
testgo:
	go test -v ./...
