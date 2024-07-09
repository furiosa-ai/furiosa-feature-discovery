FROM libfuriosa-kubernetes:latest

RUN apt-get update -qq && apt-get install -qq libssl-dev ca-certificates

ADD furiosa-feature-discovery /opt/bin/furiosa-feature-discovery