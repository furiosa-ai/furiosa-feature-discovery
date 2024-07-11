ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

all: build

clean:
	cargo clean

build:
	cargo build --release

test:
	cargo test