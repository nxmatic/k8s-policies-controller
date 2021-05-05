FROM debian:buster-slim

COPY dist/linux_linux_amd64/k8s-policies-controller /

ENTRYPOINT ["/k8s-policies-controller"]
