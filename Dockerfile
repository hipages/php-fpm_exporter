FROM golang:1.20.4 as builder

WORKDIR /go/src/github.com/hipages/php-fpm_exporter
COPY . .
RUN go mod download
RUN make build
RUN cp php-fpm_exporter /bin/php-fpm_exporter


FROM scratch as scratch
COPY --from=builder /bin/php-fpm_exporter /bin/php-fpm_exporter
EXPOSE     9253
ENTRYPOINT [ "/bin/php-fpm_exporter", "server" ]


FROM quay.io/sysdig/sysdig-mini-ubi9:1.2.0 as ubi
COPY --from=builder /bin/php-fpm_exporter /bin/php-fpm_exporter
EXPOSE     9253
ENTRYPOINT [ "/bin/php-fpm_exporter", "server" ]