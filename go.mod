module github.com/hipages/php-fpm_exporter

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/gosuri/uitable v0.0.4
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v1.19.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.9.0
	github.com/tomasen/fcgi_client v0.0.0-20180423082037-2bb3d819fd19
)

go 1.13

replace github.com/tomasen/fcgi_client => github.com/kanocz/fcgi_client v0.0.0-20210113082628-fff85c8adfb7
