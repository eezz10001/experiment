FROM  alpine:3.12.0
WORKDIR /
USER 65532:65532
COPY ./c /
COPY ./config/deploy/config /config
ENTRYPOINT ["/c"]
