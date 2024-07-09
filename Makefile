check-tag:
ifndef DOCKER_TAG
	$(error "DOCKER_TAG is not set")
endif

all: bake

clean:
	cargo clean

build:
	cargo build --release

bake: build
	docker build --no-cache -t ghcr.io/furiosa-ai/furiosa-feature-discovery ./target/release -f Dockerfile

test:
	cargo test