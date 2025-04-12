ARG base_image=cgr.dev/chainguard/wolfi-base
ARG builder_image=cgr.dev/chainguard/go

ARG TARGETOS
ARG TARGETARCH
ARG GOAMD64=v3
ARG GOARM64=v8.2

FROM --platform=$BUILDPLATFORM ${builder_image} AS builder
WORKDIR /concourse/time-resource
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOAMD64=${GOAMD64} GOARM64=${GOARM64} go build -o /assets/out github.com/concourse/time-resource/out
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOAMD64=${GOAMD64} GOARM64=${GOARM64} go build -o /assets/in github.com/concourse/time-resource/in
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOAMD64=${GOAMD64} GOARM64=${GOARM64} go build -o /assets/check github.com/concourse/time-resource/check

RUN set -e; for pkg in $(go list ./...); do \
	  go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM ${base_image} AS resource
COPY --from=builder /assets /opt/resource

FROM resource AS tests
COPY --from=builder /tests /tests
RUN set -e; for test in /tests/*.test; do \
      $test; \
    done

FROM resource
