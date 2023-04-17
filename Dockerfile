FROM centos:centos7.9.2009
WORKDIR /
USER 65532:65532
COPY ./c /
COPY ./config/deploy/config /config
ENTRYPOINT ["/c"]
