ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

all: bake

submodule_init:
	git submodule init
	git submodule update --init --recursive

submodule_update:
	git submodule update --remote --merge

clean:
	cargo clean

build: submodule_init
	cargo build --release

.PHONY: image
image:
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64

.PHONY: image-no-cache
image-no-cache:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery:devel --progress=plain --platform=linux/amd64

test:
	cargo test
