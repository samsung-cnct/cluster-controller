FROM golang:1.8

ENV GOBIN /go/bin

# build directories
RUN mkdir /app
RUN mkdir /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app

# Go dep!
RUN go get -u github.com/golang/dep/...
RUN dep ensure

# Build my app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM ubuntu:16.04

ARG JUJU_USER=ubuntu

ENV JUJU_DATA /home/$JUJU_USER/.local/share/juju
ENV JUJU_REPOSITORY /home/$JUJU_USER/charms

RUN useradd -m $JUJU_USER -s /bin/bash

RUN mkdir -p $JUJU_DATA
RUN mkdir -p $JUJU_REPOSITORY

RUN chown -R $JUJU_USER:$JUJU_USER /home/$JUJU_USER/.local
RUN chown -R $JUJU_USER:$JUJU_USER $JUJU_REPOSITORY

VOLUME [ "$JUJU_DATA", "$JUJU_REPOSITORY" ]

COPY scripts/setup.sh /setup.sh
RUN /setup.sh

ENV GOVERSION 1.10.1
ENV GOROOT /opt/go
ENV GOPATH /root/.go

RUN cd /opt && wget https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOVERSION}.linux-amd64.tar.gz && rm go${GOVERSION}.linux-amd64.tar.gz && \
    ln -s /opt/go/bin/go /usr/bin/ && \
    mkdir $GOPATH

WORKDIR /root/
RUN mkdir /bin/controller
RUN mkdir /bin/gogo-juju

# --from=0 pulls in the artifacts from the first build stage
COPY --from=0 /go/src/app /bin/controller

# USER $JUJU_USER
# gogo juju
RUN go get gopkg.in/yaml.v2
RUN git clone https://github.com/DStorck/gogo-juju.git /bin/gogo-juju

# ENTRYPOINT /bin/controller/main -kubeconfig=""
CMD ["/bin/controller/main"]