
export GO111MODULE=on

.PHONY: test
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/crd-to-openapischema github.com/replicatedhq/crd-to-openapischema/cmd/crd-to-openapischema

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

