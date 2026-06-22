ARG BASE_IMAGE=asia-northeast3-docker.pkg.dev/next-gen-infra/furiosa-ai/libfuriosa-kubernetes:v2026.1.1
FROM $BASE_IMAGE as build

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=${GITHUB_TOKEN}
RUN git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

RUN apt-get update && apt-get install -y libssl-dev clang

# Build furiosa-feature-discovery
WORKDIR /tmp
COPY . /tmp

RUN make build

ARG TARGETARCH
RUN set -e; \
    case "$TARGETARCH" in \
        amd64) libDir='x86_64-linux-gnu' ;; \
        arm64) libDir='aarch64-linux-gnu' ;; \
        *) echo >&2 "unsupported architecture: $TARGETARCH"; exit 1 ;; \
    esac; \
    mkdir -p /staging/usr/lib/$libDir; \
    cp /usr/lib/$libDir/libfuriosa_smi.so /staging/usr/lib/$libDir/libfuriosa_smi.so; \
    cp /usr/lib/$libDir/libgcc_s.so.1     /staging/usr/lib/$libDir/libgcc_s.so.1

FROM gcr.io/distroless/base-debian12:latest

# Copy binary file
COPY --from=build /tmp/target/release/furiosa-feature-discovery /opt/bin/furiosa-feature-discovery

# Below dynamic libraries are required due to `furiosa-smi` and Rust dependencies.
COPY --from=build /staging/ /

WORKDIR /opt/bin
CMD ["./furiosa-feature-discovery"]
