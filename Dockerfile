FROM alpine:3.5
LABEL maintainer "hipages DevOps Team <syd-team-devops@hipagesgroup.com.au>"

COPY php-fpm_exporter /bin/php-fpm_exporter

EXPOSE     9253
ENTRYPOINT [ "/bin/php-fpm_exporter", "server" ]
