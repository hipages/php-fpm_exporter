<?php
    $status = opcache_get_status($get_scripts = false);
    if ($status) {
        echo (json_encode($status, JSON_PRETTY_PRINT));
    } else {
        echo '{"opcache_enabled": false}';
    }
?>