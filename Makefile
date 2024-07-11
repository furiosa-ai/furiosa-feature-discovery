all: build

clean:
	cargo clean

build:
	cargo build --release

test:
	cargo test