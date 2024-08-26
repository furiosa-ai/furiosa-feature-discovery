ARG BASE_IMAGE=registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest
FROM $BASE_IMAGE as build

RUN apt-get update && apt-get install -y libssl-dev clang

# Build furiosa-feature-discovery
WORKDIR /tmp
COPY . /tmp

RUN make build-no-submodule-init

FROM gcr.io/distroless/base-debian12:latest

# Copy binary file
COPY --from=build /tmp/target/release/furiosa-feature-discovery /opt/bin/furiosa-feature-discovery

# Below dynamic libraries are required due to `furiosa-smi` and Rust dependencies.
COPY --from=build /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcc_s.so.1 /usr/lib/x86_64-linux-gnu/libgcc_s.so.1

WORKDIR /opt/bin
CMD ["./furiosa-feature-discovery"]
