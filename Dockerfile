FROM  centos:7
WORKDIR /
USER 65532:65532


RUN mkdir /root/.kube -p
COPY ./config /root/.kube/config
COPY ./c /c

