FROM golang:1.13-stretch

WORKDIR /build
COPY . .
RUN make gobuild

FROM ubuntu:18.04

WORKDIR /app

RUN mkdir /data

COPY --from=0 /build/registry .
COPY api/spec.yml spec.yml
COPY configs/config.yml config.yml

ENTRYPOINT [ "./registry" ]