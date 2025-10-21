<?php

//
// Reports status, statistics and configuration directives
// from OPcache in Prometheus format.
//
// Copyright 2017 Kumina, https://kumina.nl/
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
header('Content-Type: text/plain; version=0.0.5');

// Iterate and print in the Prometheus format
function dump_as_prometheus_metric($root, $prefix)
{
    if (is_numeric($root)) {
        echo "$prefix $root\n";
    } elseif (is_bool($root)) {
        if ($root) {
            print "$prefix 1\n";
        } else {
            print "$prefix 0\n";
        }
    } elseif (is_string($root)) {
        // Skip, as Prometheus doesn't support string values.
    } elseif (is_array($root)) {
        foreach ($root as $key => $value) {
            dump_as_prometheus_metric($value, $prefix.'_'.str_replace('.', '_', $key));
        }
    } else {
        var_dump($root);
        die('Encountered unsupported value type');
    }
}

// Report status & statistics
dump_as_prometheus_metric(opcache_get_status(false), 'opcache_status');

// Report configuration directives
dump_as_prometheus_metric(opcache_get_configuration(), 'opcache_configuration');
