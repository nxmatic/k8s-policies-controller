FROM golang:1.15-buster AS build

ENV VERSION=latest
ENV GOBIN=$GOPATH/bin

ADD . /src/k8s-policies-controller

WORKDIR /src/k8s-policies-controller

RUN make build

FROM debian:buster-slim

COPY --from=build /src/k8s-policies-controller/k8s-policies-controller /k8s-policies-controller

ENTRYPOINT ["/k8s-policies-controller"]
