# php-fpm_exporter

A prometheus exporter for PHP-FPM.
The exporter connects directly to PHP-FPM and exports the metrics via HTTP.

A webserver such as NGINX or Apache is **NOT** needed!

## Features

* Export single or multiple pools
* Export to CLI as text or JSON
* Connects directly to PHP-FPM via TCP or Socket

## Usage

* php-fpm_exporter get --phpfpm.scrape-uri 127.0.0.1:9000,127.0.0.1:9001,[...]
* php-fpm_exporter server --phpfpm.scrape-uri 127.0.0.1:9000,127.0.0.1:9001,[...]

## Metrics collected

| Metric | Type |
|--------|------|



## TODO

- [ ] Test with unix socket