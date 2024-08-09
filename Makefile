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

build:
	submodule_init
	cargo build --release

bake: build
	docker build --no-cache -t registry.furiosa.ai/furiosa/furiosa-feature-discovery . -f Dockerfile

test:
	cargo test
