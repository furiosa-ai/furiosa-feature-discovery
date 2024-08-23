BASE_IMAGE := registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

.PHONY: fmt
fmt:
	cargo fmt --all
	# cargo machete # machete is not compatible with toolchain of this repo.
	cargo sort --grouped --workspace

.PHONY: clippy
clippy:
	cargo fmt --all --check && cargo -q clippy --all-targets -- -D rust_2018_idioms -D warnings

.PHONY: submodule_init
submodule_init:
	git submodule init
	git submodule update --init --recursive

.PHONY: submodule_update
submodule_update:
	git submodule update --remote --merge

.PHONY: clean
clean:
	cargo clean

.PHONY: build
build: submodule_init
	cargo build --release

.PHONY: image
image:
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-no-cache
image-no-cache:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-rel
image-rel:
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:latest --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: image-no-cache-rel
image-no-cache-rel:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:latest --progress=plain --platform=linux/amd64 --build-arg BASE_IMAGE=$(BASE_IMAGE)

.PHONY: test
test:
	cargo test
