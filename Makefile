ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

all: bake

clean:
	cargo clean

build:
	cargo build --release

bake: build
	docker build --no-cache -t ghcr.io/furiosa-ai/furiosa-feature-discovery . -f Dockerfile

test:
	cargo test
