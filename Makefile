BASE_IMAGE := registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

ifndef GITHUB_TOKEN
$(error GITHUB_TOKEN is not set. Please set the GITHUB_TOKEN environment variable)

ifeq ($(shell uname -s),Darwin)
    CGO_CFLAGS := "-I/usr/local/include"
    CGO_LDFLAGS := "-L/usr/local/lib"
endif

.PHONY: fmt
fmt:
	cargo fmt --all
	# cargo machete # machete is not compatible with toolchain of this repo.
	cargo sort --grouped --workspace

.PHONY: clippy
clippy:
	cargo fmt --all --check && cargo -q clippy --all-targets -- -D rust_2018_idioms -D warnings

.PHONY: clean
clean:
	cargo clean

.PHONY: build
build:
	cargo build --release

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

.PHONY:e2e-feature-discovery
e2e-feature-discovery:
	# build container image
	# run e2e test framework
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) ginkgo ./e2e