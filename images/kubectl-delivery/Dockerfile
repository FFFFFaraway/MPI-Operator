FROM golang:1.13.6 AS build

# Install kubectl
ENV K8S_VERSION v1.21.0
RUN apt-get install wget
RUN wget -q https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /bin/kubectl

FROM alpine:3.11.6
COPY --from=build /bin/kubectl /bin/kubectl
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["cp /bin/kubectl /opt/kube/kubectl"]
