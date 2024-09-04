BASE_IMAGE := registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

ifndef GITHUB_TOKEN
$(error GITHUB_TOKEN is not set. Please set the GITHUB_TOKEN environment variable)
endif

CGO_CFLAGS := -I/usr/local/include
CGO_LDFLAGS := -L/usr/local/lib

LABELS_TO_REMOVE := furiosa.ai/driver.version furiosa.ai/driver.version.major furiosa.ai/driver.version.minor furiosa.ai/driver.version.patch furiosa.ai/driver.version.metadata furiosa.ai/npu.count furiosa.ai/npu.family furiosa.ai/npu.product

NODES := $(shell kubectl get nodes -o name | sed 's/node\///')

E2E_TEST_IMAGE_REGISTRY := registry.corp.furiosa.ai/furiosa
E2E_TEST_IMAGE_NAME := furiosa-feature-discovery
E2E_TEST_IMAGE_TAG := latest

.PHONY: fmt
fmt: fmt-rs fmt-go

.PHONY: fmt-rs
fmt-rs:
	cargo fmt --all
	# cargo machete # machete is not compatible with toolchain of this repo.
	cargo sort --grouped --workspace

.PHONY: fmt-go
fmt-go:
	go fmt ./...

.PHONY: lint
lint: clippy lint-go

.PHONY: clippy
clippy:
	cargo fmt --all --check && cargo -q clippy --all-targets -- -D rust_2018_idioms -D warnings

.PHONY: lint-go
lint-go:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) golangci-lint run --timeout=30m

.PHONY: vet
vet:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go vet -v ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: clean
clean:
	cargo clean

.PHONY: build
build:
	cargo build --release

.PHONY: build-go
build-go:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go build ./...

.PHONY: image
image:
	docker build . --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-no-cache
image-no-cache:
	docker build . --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-rel
image-rel:
	docker build . --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:latest --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-no-cache-rel
image-no-cache-rel:
	docker build . --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:latest --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: test
test:
	cargo test

.PHONY: clean-labels
clean-labels:
	@echo "Labels to remove: $(LABELS_TO_REMOVE)"
	@for node in $(NODES); do \
		echo "Processing node: $$node"; \
		CURRENT_LABELS=$$(kubectl get node $$node --show-labels | grep -oP '(?<=,|^)\K[^=]+'); \
		for label in $(LABELS_TO_REMOVE); do \
			if echo $$CURRENT_LABELS | grep -q "$$label"; then \
				echo "Removing label $$label from node $$node"; \
				kubectl label node $$node $$label-; \
			else \
				echo "Label $$label not found on node $$node, skipping"; \
			fi \
		done \
	done

.PHONY:e2e
e2e: clean-labels
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) E2E_TEST_IMAGE_REGISTRY=$(E2E_TEST_IMAGE_REGISTRY) E2E_TEST_IMAGE_NAME=$(E2E_TEST_IMAGE_NAME) E2E_TEST_IMAGE_TAG=$(E2E_TEST_IMAGE_TAG) ginkgo ./e2e
