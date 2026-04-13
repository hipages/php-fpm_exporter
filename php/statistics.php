<?php

/**
 * OPcache statistics script for php-fpm_exporter.
 *
 * This script must be deployed inside the PHP-FPM container at the path
 * configured via --phpfpm.opcache-script (default: /app/statistics.php).
 * It is executed by the exporter via FastCGI on every scrape and must
 * output a JSON object compatible with the phpfpm_opcache_* metrics.
 */

header('Content-Type: application/json');

if (!extension_loaded('Zend OPcache')) {
    echo json_encode([
        'opcache_enabled'    => false,
        'cache_full'         => false,
        'restart_pending'    => false,
        'restart_in_progress' => false,
        'memory_usage' => [
            'used_memory'               => 0,
            'free_memory'               => 0,
            'wasted_memory'             => 0,
            'current_wasted_percentage' => 0.0,
        ],
        'interned_strings_usage' => [
            'buffer_size'      => 0,
            'used_memory'      => 0,
            'free_memory'      => 0,
            'number_of_strings' => 0,
        ],
        'opcache_statistics' => [
            'num_cached_scripts'  => 0,
            'num_cached_keys'     => 0,
            'max_cached_keys'     => 0,
            'hits'                => 0,
            'start_time'          => 0,
            'last_restart_time'   => 0,
            'oom_restarts'        => 0,
            'hash_restarts'       => 0,
            'manual_restarts'     => 0,
            'misses'              => 0,
            'blacklist_misses'    => 0,
            'blacklist_miss_ratio' => 0.0,
            'opcache_hit_rate'    => 0.0,
        ],
    ]);
    exit;
}

// Exclude per-script details (second argument false) to keep the
// response small and avoid variable-length output.
$status = opcache_get_status(false);

if ($status === false) {
    // OPcache is loaded but disabled at runtime.
    echo json_encode([
        'opcache_enabled'    => false,
        'cache_full'         => false,
        'restart_pending'    => false,
        'restart_in_progress' => false,
        'memory_usage' => [
            'used_memory'               => 0,
            'free_memory'               => 0,
            'wasted_memory'             => 0,
            'current_wasted_percentage' => 0.0,
        ],
        'interned_strings_usage' => [
            'buffer_size'      => 0,
            'used_memory'      => 0,
            'free_memory'      => 0,
            'number_of_strings' => 0,
        ],
        'opcache_statistics' => [
            'num_cached_scripts'  => 0,
            'num_cached_keys'     => 0,
            'max_cached_keys'     => 0,
            'hits'                => 0,
            'start_time'          => 0,
            'last_restart_time'   => 0,
            'oom_restarts'        => 0,
            'hash_restarts'       => 0,
            'manual_restarts'     => 0,
            'misses'              => 0,
            'blacklist_misses'    => 0,
            'blacklist_miss_ratio' => 0.0,
            'opcache_hit_rate'    => 0.0,
        ],
    ]);
    exit;
}

echo json_encode($status);
