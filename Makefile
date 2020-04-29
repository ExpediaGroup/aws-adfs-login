test:
	go fmt ./...
	GO111MODULE=on go test -v ./...

build: test
	GO111MODULE=on go build -v ./pkg/client

build_release_artifacts: build
	@[ "${filename}" ] || (echo ">> filename is not set. Should be of format v<major>.<minor>.<patch>"; exit 1)
	zip -r --include '*.go' 'go.mod' 'go.sum' 'LICENSE' 'NOTICE' 'README.md' @ $(filename).zip . -x '*_test.go'
	tar --exclude='*_test.go' -cvf $(filename).tar.gz `find . | egrep ".*\.go|go\.mod|go\.sum|LICENSE|NOTICE|README\.md"`
