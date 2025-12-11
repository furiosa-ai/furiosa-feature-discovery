ARG BASE_IMAGE=asia-northeast3-docker.pkg.dev/next-gen-infra/furiosa-ai/libfuriosa-kubernetes:v0.3.1
FROM $BASE_IMAGE as build

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=${GITHUB_TOKEN}
RUN git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

RUN apt-get update && apt-get install -y libssl-dev clang

# Build furiosa-feature-discovery
WORKDIR /tmp
COPY . /tmp

RUN make build

FROM gcr.io/distroless/base-debian12:latest

# Copy binary file
COPY --from=build /tmp/target/release/furiosa-feature-discovery /opt/bin/furiosa-feature-discovery

# Below dynamic libraries are required due to `furiosa-smi` and Rust dependencies.
COPY --from=build /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcc_s.so.1 /usr/lib/x86_64-linux-gnu/libgcc_s.so.1

WORKDIR /opt/bin
CMD ["./furiosa-feature-discovery"]
