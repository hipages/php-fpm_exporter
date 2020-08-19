<?php

header('Content-Type: application/json; version=0.0.4');

$metrics = array();

// Check if apcu is enabled
if (function_exists('apcu_cache_info')) {
  $metrics['apcu'] = apcu_cache_info();
}

// Check if apc is enabled
if (function_exists('apc_sma_info')) {
  $metrics['apc'] = apc_sma_info();
}

// Check if opcache is enabled
if (function_exists('opcache_get_status')) {
  $metrics['opcache'] = opcache_get_status();
}

if (function_exists('getrusage')) {
  $metrics['getrusage'] = getrusage();
}

$metrics['realpath'] = array('entries' =>  realpath_cache_get(), 'size' => realpath_cache_size());

print_r(json_encode($metrics,  JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));
?>