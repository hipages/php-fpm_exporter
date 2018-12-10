FROM golang:1.11 as build

WORKDIR /app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:3.5
LABEL maintainer "hipages DevOps Team <syd-team-devops@hipagesgroup.com.au>"

COPY --from=build /go/bin/php-fpm_exporter /

EXPOSE     9253
ENTRYPOINT [ "/php-fpm_exporter", "server" ]
