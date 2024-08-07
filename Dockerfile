FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:latest as smi

FROM rust:bookworm as build

# Install dependencies
RUN apt-get update && \
    apt-get install -y \
    build-essential \
    clang

COPY --from=smi /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=smi /usr/include/furiosa/furiosa_smi.h /usr/include/furiosa/furiosa_smi.h
RUN ldconfig

# Build furiosa-feature-discovery
WORKDIR /tmp
COPY . /tmp

RUN make build

FROM ubuntu:22.04

COPY --from=smi /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=smi /usr/include/furiosa/furiosa_smi.h /usr/include/furiosa/furiosa_smi.h
RUN ldconfig

COPY --from=build /tmp/target/release/furiosa-feature-discovery /opt/bin/furiosa-feature-discovery
WORKDIR /opt/bin
