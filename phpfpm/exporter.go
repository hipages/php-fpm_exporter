// Copyright © 2018 Enrico Stahn <enrico.stahn@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package phpfpm provides convenient access to PHP-FPM pool data
package phpfpm

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	hashids "github.com/speps/go-hashids"
)

const (
	namespace = "phpfpm"
)

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	mutex       sync.Mutex
	PoolManager PoolManager
	Opcache     Opcache

	CountProcessState  bool
	EnableOpcacheStats bool

	up                       *prometheus.Desc
	scrapeFailues            *prometheus.Desc
	startSince               *prometheus.Desc
	acceptedConnections      *prometheus.Desc
	listenQueue              *prometheus.Desc
	maxListenQueue           *prometheus.Desc
	listenQueueLength        *prometheus.Desc
	idleProcesses            *prometheus.Desc
	activeProcesses          *prometheus.Desc
	totalProcesses           *prometheus.Desc
	maxActiveProcesses       *prometheus.Desc
	maxChildrenReached       *prometheus.Desc
	slowRequests             *prometheus.Desc
	processRequests          *prometheus.Desc
	processLastRequestMemory *prometheus.Desc
	processLastRequestCPU    *prometheus.Desc
	processRequestDuration   *prometheus.Desc
	processState             *prometheus.Desc

	opcacheEnabled   *prometheus.Desc
	opcacheCacheFull *prometheus.Desc

	opcacheMemoryUsed      *prometheus.Desc
	opcacheMemoryFree      *prometheus.Desc
	opcacheMemoryWasted    *prometheus.Desc
	opcacheMemoryWastedPct *prometheus.Desc

	opcacheInternedStringBufferSize *prometheus.Desc
	opcacheInternedStringMemoryUsed *prometheus.Desc
	opcacheInternedStringMemoryFree *prometheus.Desc
	opcacheInternedStringCount      *prometheus.Desc

	opcacheCachedScripts *prometheus.Desc
	opcacheCachedKeys    *prometheus.Desc
	opcacheCachedKeysMax *prometheus.Desc
	opcacheCacheHits     *prometheus.Desc
	opcacheCacheMisses   *prometheus.Desc
	opcacheCacheHitRate  *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,
		Opcache:     Opcache{},

		CountProcessState: false,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could PHP-FPM be reached?",
			[]string{"pool"},
			nil),

		scrapeFailues: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_failures"),
			"The number of failures scraping from PHP-FPM.",
			[]string{"pool"},
			nil),

		startSince: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_since"),
			"The number of seconds since FPM has started.",
			[]string{"pool"},
			nil),

		acceptedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "accepted_connections"),
			"The number of requests accepted by the pool.",
			[]string{"pool"},
			nil),

		listenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue"),
			"The number of requests in the queue of pending connections.",
			[]string{"pool"},
			nil),

		maxListenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_listen_queue"),
			"The maximum number of requests in the queue of pending connections since FPM has started.",
			[]string{"pool"},
			nil),

		listenQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue_length"),
			"The size of the socket queue of pending connections.",
			[]string{"pool"},
			nil),

		idleProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "idle_processes"),
			"The number of idle processes.",
			[]string{"pool"},
			nil),

		activeProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_processes"),
			"The number of active processes.",
			[]string{"pool"},
			nil),

		totalProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_processes"),
			"The number of idle + active processes.",
			[]string{"pool"},
			nil),

		maxActiveProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_active_processes"),
			"The maximum number of active processes since FPM has started.",
			[]string{"pool"},
			nil),

		maxChildrenReached: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_children_reached"),
			"The number of times, the process limit has been reached, when pm tries to start more children (works only for pm 'dynamic' and 'ondemand').",
			[]string{"pool"},
			nil),

		slowRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "slow_requests"),
			"The number of requests that exceeded your 'request_slowlog_timeout' value.",
			[]string{"pool"},
			nil),

		processRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_requests"),
			"The number of requests the process has served.",
			[]string{"pool", "pid_hash"},
			nil),

		processLastRequestMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_memory"),
			"The max amount of memory the last request consumed.",
			[]string{"pool", "pid_hash"},
			nil),

		processLastRequestCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_cpu"),
			"The %cpu the last request consumed.",
			[]string{"pool", "pid_hash"},
			nil),

		processRequestDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_duration"),
			"The duration in microseconds of the requests.",
			[]string{"pool", "pid_hash"},
			nil),

		processState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_state"),
			"The state of the process (Idle, Running, ...).",
			[]string{"pool", "pid_hash", "state"},
			nil),

		opcacheEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "enabled"),
			"Is PHP Opcache enabled?",
			nil,
			nil),

		opcacheCacheFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cache_full"),
			"Is PHP Opcache cache full?",
			nil,
			nil),

		opcacheMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "memory_used_bytes"),
			"PHP Opcache cache memory used in bytes",
			nil,
			nil),

		opcacheMemoryFree: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "memory_free_bytes"),
			"PHP Opcache cache memory free in bytes",
			nil,
			nil),

		opcacheMemoryWasted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "memory_wasted_bytes"),
			"PHP Opcache cache memory wasted in bytes",
			nil,
			nil),

		opcacheMemoryWastedPct: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "memory_wasted_percent"),
			"Percent of PHP Opcache cache memory wasted",
			nil,
			nil),

		opcacheInternedStringBufferSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "interned_string_buffer_bytes"),
			"PHP Opcache interned string buffer size",
			nil,
			nil),

		opcacheInternedStringMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "interned_string_memory_used_bytes"),
			"PHP Opcache interned string memory used",
			nil,
			nil),

		opcacheInternedStringMemoryFree: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "interned_string_memory_free_bytes"),
			"PHP Opcache interned string memory free",
			nil,
			nil),

		opcacheInternedStringCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "interned_string_total"),
			"PHP Opcache total number of interned strings",
			nil,
			nil),

		opcacheCachedScripts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cached_scripts_total"),
			"PHP Opcache total number of cached scripts",
			nil,
			nil),

		opcacheCachedKeys: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cached_keys_total"),
			"PHP Opcache total number of cached keys",
			nil,
			nil),

		opcacheCachedKeysMax: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cached_keys_max_total"),
			"PHP Opcache maximum number of cached keys",
			nil,
			nil),

		opcacheCacheHits: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cache_hits_total"),
			"PHP Opcache number of cache hits",
			nil,
			nil),

		opcacheCacheMisses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cache_misses_total"),
			"PHP Opcache number of cache misses",
			nil,
			nil),

		opcacheCacheHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "opcache", "cache_hit_rate_percent"),
			"PHP Opcache rate of cache hits",
			nil,
			nil),
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.PoolManager.Update()

	// Use the first available pool to fetch the opcache statistics
	// Opcache stats are global to the FPM process so no need to fetch these for every pool
	if e.EnableOpcacheStats && len(e.PoolManager.Pools) >= 1 {
		p := e.PoolManager.Pools[0]
		oc := &e.Opcache
		log.Debugf("Using pool %v for opcache statistics", p.Name)
		err := oc.Update(p)
		if err != nil {
			log.Errorf("Error updating opcache statistics: %v", err)
		}
	}

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.scrapeFailues, prometheus.CounterValue, float64(pool.ScrapeFailures), pool.Name)

		if pool.ScrapeError != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, pool.Name)
			log.Errorf("Error scraping PHP-FPM: %v", pool.ScrapeError)
			continue
		}

		active, idle, total := CountProcessState(pool.Processes)
		if !e.CountProcessState && (active != pool.ActiveProcesses || idle != pool.IdleProcesses) {
			log.Error("Inconsistent active and idle processes reported. Set `--fix-process-count` to have this calculated by php-fpm_exporter instead.")
		}

		if !e.CountProcessState {
			active = pool.ActiveProcesses
			idle = pool.IdleProcesses
			total = pool.TotalProcesses
		}

		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1, pool.Name)
		ch <- prometheus.MustNewConstMetric(e.startSince, prometheus.CounterValue, float64(pool.StartSince), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.acceptedConnections, prometheus.CounterValue, float64(pool.AcceptedConnections), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueue, prometheus.GaugeValue, float64(pool.ListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxListenQueue, prometheus.CounterValue, float64(pool.MaxListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueueLength, prometheus.GaugeValue, float64(pool.ListenQueueLength), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.idleProcesses, prometheus.GaugeValue, float64(idle), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.activeProcesses, prometheus.GaugeValue, float64(active), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.totalProcesses, prometheus.GaugeValue, float64(total), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxActiveProcesses, prometheus.CounterValue, float64(pool.MaxActiveProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxChildrenReached, prometheus.CounterValue, float64(pool.MaxChildrenReached), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.slowRequests, prometheus.CounterValue, float64(pool.SlowRequests), pool.Name)

		for _, process := range pool.Processes {
			pidHash := calculateProcessHash(process)
			ch <- prometheus.MustNewConstMetric(e.processState, prometheus.GaugeValue, 1, pool.Name, pidHash, process.State)
			ch <- prometheus.MustNewConstMetric(e.processRequests, prometheus.CounterValue, float64(process.Requests), pool.Name, pidHash)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestMemory, prometheus.GaugeValue, float64(process.LastRequestMemory), pool.Name, pidHash)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestCPU, prometheus.GaugeValue, process.LastRequestCPU, pool.Name, pidHash)
			ch <- prometheus.MustNewConstMetric(e.processRequestDuration, prometheus.GaugeValue, float64(process.RequestDuration), pool.Name, pidHash)
		}
	}

	if e.EnableOpcacheStats {
		ch <- prometheus.MustNewConstMetric(e.opcacheEnabled, prometheus.GaugeValue, (map[bool]float64{true: 1, false: 0})[e.Opcache.OpcacheEnabled])
		// If Opcache is reported as enabled by the opcache stats PHP script, include the following metrics
		if e.Opcache.OpcacheEnabled {
			ch <- prometheus.MustNewConstMetric(e.opcacheCacheFull, prometheus.GaugeValue, (map[bool]float64{true: 1, false: 0})[e.Opcache.CacheFull])
			ch <- prometheus.MustNewConstMetric(e.opcacheMemoryUsed, prometheus.GaugeValue, float64(e.Opcache.MemoryUsage.UsedMemory))
			ch <- prometheus.MustNewConstMetric(e.opcacheMemoryFree, prometheus.GaugeValue, float64(e.Opcache.MemoryUsage.FreeMemory))
			ch <- prometheus.MustNewConstMetric(e.opcacheMemoryWasted, prometheus.GaugeValue, float64(e.Opcache.MemoryUsage.WastedMemory))
			ch <- prometheus.MustNewConstMetric(e.opcacheMemoryWastedPct, prometheus.GaugeValue, float64(e.Opcache.MemoryUsage.CurrentWastedPercentage))

			ch <- prometheus.MustNewConstMetric(e.opcacheInternedStringBufferSize, prometheus.GaugeValue, float64(e.Opcache.InternedStringsUsage.BufferSize))
			ch <- prometheus.MustNewConstMetric(e.opcacheInternedStringMemoryUsed, prometheus.GaugeValue, float64(e.Opcache.InternedStringsUsage.UsedMemory))
			ch <- prometheus.MustNewConstMetric(e.opcacheInternedStringMemoryFree, prometheus.GaugeValue, float64(e.Opcache.InternedStringsUsage.FreeMemory))
			ch <- prometheus.MustNewConstMetric(e.opcacheInternedStringCount, prometheus.GaugeValue, float64(e.Opcache.InternedStringsUsage.NumberOfStrings))

			ch <- prometheus.MustNewConstMetric(e.opcacheCachedScripts, prometheus.GaugeValue, float64(e.Opcache.OpcacheStatistics.NumCachedScripts))
			ch <- prometheus.MustNewConstMetric(e.opcacheCachedKeys, prometheus.GaugeValue, float64(e.Opcache.OpcacheStatistics.NumCachedKeys))
			ch <- prometheus.MustNewConstMetric(e.opcacheCachedKeysMax, prometheus.CounterValue, float64(e.Opcache.OpcacheStatistics.MaxCachedKeys))
			ch <- prometheus.MustNewConstMetric(e.opcacheCacheHits, prometheus.CounterValue, float64(e.Opcache.OpcacheStatistics.Hits))
			ch <- prometheus.MustNewConstMetric(e.opcacheCacheMisses, prometheus.CounterValue, float64(e.Opcache.OpcacheStatistics.Misses))
			ch <- prometheus.MustNewConstMetric(e.opcacheCacheHitRate, prometheus.GaugeValue, float64(e.Opcache.OpcacheStatistics.OpcacheHitRate))
		}
	}
}

// Describe exposes the metric description to Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.startSince
	ch <- e.acceptedConnections
	ch <- e.listenQueue
	ch <- e.maxListenQueue
	ch <- e.listenQueueLength
	ch <- e.idleProcesses
	ch <- e.activeProcesses
	ch <- e.totalProcesses
	ch <- e.maxActiveProcesses
	ch <- e.maxChildrenReached
	ch <- e.slowRequests
	ch <- e.processState
	ch <- e.processRequests
	ch <- e.processLastRequestMemory
	ch <- e.processLastRequestCPU
	ch <- e.processRequestDuration
	ch <- e.opcacheEnabled
	ch <- e.opcacheCacheFull
	ch <- e.opcacheMemoryUsed
	ch <- e.opcacheMemoryFree
	ch <- e.opcacheMemoryWasted
	ch <- e.opcacheMemoryWastedPct
	ch <- e.opcacheInternedStringBufferSize
	ch <- e.opcacheInternedStringMemoryUsed
	ch <- e.opcacheInternedStringMemoryFree
	ch <- e.opcacheInternedStringCount
	ch <- e.opcacheCachedScripts
	ch <- e.opcacheCachedKeys
	ch <- e.opcacheCachedKeysMax
	ch <- e.opcacheCacheHits
	ch <- e.opcacheCacheMisses
	ch <- e.opcacheCacheHitRate
}

// calculateProcessHash generates a unique identifier for a process to ensure uniqueness across multiple systems/containers
func calculateProcessHash(pp PoolProcess) string {
	hd := hashids.NewData()
	hd.Salt = "php-fpm_exporter"
	hd.MinLength = 12
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(pp.StartTime), int(pp.PID)})

	return e
}
