# Build the binary
FROM docker.io/library/golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY phpfpm/ phpfpm/

# Build
RUN CGO_ENABLED=0 go build -a -o php-fpm_exporter main.go

FROM docker.io/library/alpine:3.15.0

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

COPY --from=builder /workspace/php-fpm_exporter .

EXPOSE     9253
ENTRYPOINT [ "/php-fpm_exporter", "server" ]

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="php-fpm_exporter" \
      org.label-schema.description="A prometheus exporter for PHP-FPM." \
      org.label-schema.url="https://hipages.com.au/" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hipages/php-fpm_exporter" \
      org.label-schema.vendor="hipages" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0" \
      org.label-schema.docker.cmd="docker run -it --rm -e PHP_FPM_SCRAPE_URI=\"tcp://127.0.0.1:9000/status\" hipages/php-fpm_exporter"
