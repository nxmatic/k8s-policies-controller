FROM golang:1.15-buster AS build

ARG DOCKER_REGISTRY
ARG DOCKER_REGISTRY_ORG
ARG VERSION

ENV DOCKER_REGISTRY=$DOCKER_REGISTRY
ENV DOCKER_REGISTRY_ORG=$DOCKER_REGISTRY_ORG
ENV VERSION=$VERSION

ENV GOBIN=$GOPATH/bin

ADD . /src/k8s-policies-controller

WORKDIR /src/k8s-policies-controller

RUN make build

FROM debian:buster-slim

COPY --from=build /src/k8s-policies-controller/k8s-policies-controller /k8s-policies-controller

ENTRYPOINT ["/k8s-policies-controller"]
