#!/usr/bin/env bats

load '/usr/local/lib/bats-support/load.bash'
load '/usr/local/lib/bats-assert/load.bash'

setup () {
	docker-compose -f ./test/docker-compose-e2e.yml up -d
    # Workaround to wait for docker to start containers
    sleep 5
    #go run main.go server --phpfpm.scrape-uri tcp://127.0.0.1:9031/status,tcp://127.0.0.1:9032/status 3>&- &
}

teardown() {
	docker-compose -f ./test/docker-compose-e2e.yml stop
	docker-compose -f ./test/docker-compose-e2e.yml rm -f
	docker-compose -f ./test/docker-compose-e2e.yml down --volumes
}

@test "Should have metrics endpoint" {
    run curl -sSL http://localhost:9253/metrics
    [ "$status" -eq 0 ]
}

@test "Should have metric phpfpm_up" {
    run curl -sSL http://localhost:9253/metrics
    assert_output --partial '# TYPE phpfpm_up gauge'
}

@test "Should have scraped multiple PHP-FPM pools" {
    run curl -sSL http://localhost:9253/metrics
    assert_output --partial 'phpfpm_up{pool="www",scrape_uri="tcp://phpfpm1:9000/status"} 1'
    assert_output --partial 'phpfpm_up{pool="www",scrape_uri="tcp://phpfpm2:9000/status"} 1'
    assert_output --partial 'phpfpm_up{pool="www",scrape_uri="tcp://phpfpm3:9000/status"} 1'
}

@test "Should exit cleanly" {
    run docker-compose -f ./test/docker-compose-e2e.yml stop exporter
    docker-compose -f ./test/docker-compose-e2e.yml ps exporter | grep -q "Exit 0"
}
