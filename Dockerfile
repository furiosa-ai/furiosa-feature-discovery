FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest as build

RUN apt-get update && apt-get install -y libssl-dev clang

# Build furiosa-feature-discovery
WORKDIR /tmp
COPY . /tmp

RUN make build

FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

COPY --from=build /tmp/target/release/furiosa-feature-discovery /opt/bin/furiosa-feature-discovery
WORKDIR /opt/bin
