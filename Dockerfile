FROM ubuntu:14.04

RUN apt-get update && apt-get -y install wget

ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /tmp/go

ADD . /tmp/go/src/github.com/concourse/time-resource

RUN wget -qO- https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz | tar -C /usr/local -xzf - && \
      go build -o /tmp/resource/check github.com/concourse/time-resource/check && \
      go build -o /tmp/resource/in github.com/concourse/time-resource/in && \
      rm -rf /tmp/go /usr/local/go
