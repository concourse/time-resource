ARG base_image=ubuntu:latest
ARG builder_image=concourse/golang-builder

FROM busybox:uclibc as busybox

FROM ${builder_image} AS builder
WORKDIR /concourse/time-resource
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . /concourse/time-resource
ENV CGO_ENABLED 0
RUN go build -o /assets/out github.com/concourse/time-resource/out
RUN go build -o /assets/in github.com/concourse/time-resource/in
RUN go build -o /assets/check github.com/concourse/time-resource/check
RUN set -e; for pkg in $(go list ./...); do \
	go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM ${base_image} AS resource
USER root
COPY --from=builder /assets /opt/resource
COPY --from=busybox /bin/sh /bin/sh
COPY --from=busybox /bin/cp /bin/cp

FROM resource AS tests
COPY --from=builder /tests /tests
RUN set -e; for test in /tests/*.test; do \
	$test; \
	done

FROM resource
