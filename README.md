# php-fpm_exporter

[![CircleCI](https://circleci.com/gh/hipages/php-fpm_exporter.svg?style=svg)](https://circleci.com/gh/hipages/php-fpm_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/hipages/php-fpm_exporter)](https://goreportcard.com/report/github.com/hipages/php-fpm_exporter)
[![GoDoc](https://godoc.org/github.com/hipages/php-fpm_exporter?status.svg)](https://godoc.org/github.com/hipages/php-fpm_exporter)
[![Inline docs](http://inch-ci.org/github/hipages/php-fpm_exporter.svg?branch=master)](http://inch-ci.org/github/hipages/php-fpm_exporter)
[![Maintainability](https://api.codeclimate.com/v1/badges/52f9e1f0388e8aa38bfe/maintainability)](https://codeclimate.com/github/hipages/php-fpm_exporter/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/52f9e1f0388e8aa38bfe/test_coverage)](https://codeclimate.com/github/hipages/php-fpm_exporter/test_coverage)

A prometheus exporter for PHP-FPM.
The exporter connects directly to PHP-FPM and exports the metrics via HTTP.

A webserver such as NGINX or Apache is **NOT** needed!

## Features

* Export single or multiple pools
* Export to CLI as text or JSON
* Connects directly to PHP-FPM via TCP or Socket

## Usage

* ```php-fpm_exporter get --phpfpm.scrape-uri 127.0.0.1:9000,127.0.0.1:9001,[...]```
* ```php-fpm_exporter server --phpfpm.scrape-uri 127.0.0.1:9000,127.0.0.1:9001,[...]```

## Metrics collected

| Metric | Type |
|--------|------|



## TODO

- [ ] Test with unix socket

## Contributing

Contributions are greatly appreciated.
The maintainers actively manage the issues list, and try to highlight issues suitable for newcomers.
The project follows the typical GitHub pull request model.
See " [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/) " for more details.
Before starting any work, please either comment on an existing issue, or file a new one.