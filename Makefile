.PHONY: build
build:
	go build -o mackerel-plugin-mongodb

.PHONY: test
test: testgo metrictest

.PHONY: metrictest
metrictest: build
	go install github.com/lufia/graphitemetrictest/cmd/graphite-metric-test@latest
	./test.sh

.PHONY: testgo
testgo:
	go test -v ./...
