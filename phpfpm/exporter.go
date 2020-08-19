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
	namespace        = "phpfpm"
	opcacheNamespace = "opcache"
)

var (
	enabledDesc           = newMetric("enabled", "Is OPcache enabled.")
	cacheFullDesc         = newMetric("cache_full", "Is OPcache full.")
	restartPendingDesc    = newMetric("restart_pending", "Is restart pending.")
	restartInProgressDesc = newMetric("restart_in_progress", "Is restart in progress.")

	memoryUsageUsedMemoryDesc              = newMetric("memory_usage_used_memory", "OPcache used memory.")
	memoryUsageFreeMemoryDesc              = newMetric("memory_usage_free_memory", "OPcache free memory.")
	memoryUsageWastedMemoryDesc            = newMetric("memory_usage_wasted_memory", "OPcache wasted memory.")
	memoryUsageCurrentWastedPercentageDesc = newMetric("memory_usage_current_wasted_percentage", "OPcache current wasted percentage.")

	internedStringsUsageBufferSizeDesc     = newMetric("interned_strings_usage_buffer_size", "OPcache interned string buffer size.")
	internedStringsUsageUsedMemoryDesc     = newMetric("interned_strings_usage_used_memory", "OPcache interned string used memory.")
	internedStringsUsageUsedFreeMemory     = newMetric("interned_strings_usage_free_memory", "OPcache interned string free memory.")
	internedStringsUsageUsedNumerOfStrings = newMetric("interned_strings_usage_number_of_strings", "OPcache interned string number of strings.")

	statisticsNumCachedScripts   = newMetric("statistics_num_cached_scripts", "OPcache statistics, number of cached scripts.")
	statisticsNumCachedKeys      = newMetric("statistics_num_cached_keys", "OPcache statistics, number of cached keys.")
	statisticsMaxCachedKeys      = newMetric("statistics_max_cached_keys", "OPcache statistics, max cached keys.")
	statisticsHits               = newMetric("statistics_hits", "OPcache statistics, hits.")
	statisticsStartTime          = newMetric("statistics_start_time", "OPcache statistics, start time.")
	statisticsLastRestartTime    = newMetric("statistics_last_restart_time", "OPcache statistics, last restart time")
	statisticsOOMRestarts        = newMetric("statistics_oom_restarts", "OPcache statistics, oom restarts")
	statisticsHashRestarts       = newMetric("statistics_hash_restarts", "OPcache statistics, hash restarts")
	statisticsManualRestarts     = newMetric("statistics_manual_restarts", "OPcache statistics, manual restarts")
	statisticsMisses             = newMetric("statistics_misses", "OPcache statistics, misses")
	statisticsBlacklistMisses    = newMetric("statistics_blacklist_misses", "OPcache statistics, blacklist misses")
	statisticsBlacklistMissRatio = newMetric("statistics_blacklist_miss_ratio", "OPcache statistics, blacklist miss ratio")
	statisticsHitRate            = newMetric("statistics_hit_rate", "OPcache statistics, opcache hit rate")
)

func newMetric(metricName, metricDesc string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(opcacheNamespace, "", metricName), metricDesc, nil, nil)
}

func boolMetric(value bool) float64 {
	return map[bool]float64{true: 1, false: 0}[value]
}

func intMetric(value int64) float64 {
	return float64(value)
}

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	mutex       sync.Mutex
	PoolManager PoolManager

	CountProcessState bool

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
	opcacheEnabled           *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,

		CountProcessState: false,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could PHP-FPM be reached?",
			[]string{"pool", "scrape_uri"},
			nil),

		scrapeFailues: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_failures"),
			"The number of failures scraping from PHP-FPM.",
			[]string{"pool", "scrape_uri"},
			nil),

		startSince: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "start_since"),
			"The number of seconds since FPM has started.",
			[]string{"pool", "scrape_uri"},
			nil),

		acceptedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "accepted_connections"),
			"The number of requests accepted by the pool.",
			[]string{"pool", "scrape_uri"},
			nil),

		listenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue"),
			"The number of requests in the queue of pending connections.",
			[]string{"pool", "scrape_uri"},
			nil),

		maxListenQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_listen_queue"),
			"The maximum number of requests in the queue of pending connections since FPM has started.",
			[]string{"pool", "scrape_uri"},
			nil),

		listenQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "listen_queue_length"),
			"The size of the socket queue of pending connections.",
			[]string{"pool", "scrape_uri"},
			nil),

		idleProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "idle_processes"),
			"The number of idle processes.",
			[]string{"pool", "scrape_uri"},
			nil),

		activeProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_processes"),
			"The number of active processes.",
			[]string{"pool", "scrape_uri"},
			nil),

		totalProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_processes"),
			"The number of idle + active processes.",
			[]string{"pool", "scrape_uri"},
			nil),

		maxActiveProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_active_processes"),
			"The maximum number of active processes since FPM has started.",
			[]string{"pool", "scrape_uri"},
			nil),

		maxChildrenReached: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_children_reached"),
			"The number of times, the process limit has been reached, when pm tries to start more children (works only for pm 'dynamic' and 'ondemand').",
			[]string{"pool", "scrape_uri"},
			nil),

		slowRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "slow_requests"),
			"The number of requests that exceeded your 'request_slowlog_timeout' value.",
			[]string{"pool", "scrape_uri"},
			nil),

		processRequests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_requests"),
			"The number of requests the process has served.",
			[]string{"pool", "pid_hash", "scrape_uri"},
			nil),

		processLastRequestMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_memory"),
			"The max amount of memory the last request consumed.",
			[]string{"pool", "pid_hash", "scrape_uri"},
			nil),

		processLastRequestCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_cpu"),
			"The %cpu the last request consumed.",
			[]string{"pool", "pid_hash", "scrape_uri"},
			nil),

		processRequestDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_request_duration"),
			"The duration in microseconds of the requests.",
			[]string{"pool", "pid_hash", "scrape_uri"},
			nil),

		processState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_state"),
			"The state of the process (Idle, Running, ...).",
			[]string{"pool", "pid_hash", "state", "scrape_uri"},
			nil),
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if err := e.PoolManager.Update(); err != nil {
		log.Error(err)
	}

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.scrapeFailues, prometheus.CounterValue, float64(pool.ScrapeFailures), pool.Name, pool.ScrapeHost+pool.ScrapePath)

		if pool.ScrapeError != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, pool.Name, pool.ScrapeHost+pool.ScrapePath)
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

		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1, pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.startSince, prometheus.CounterValue, float64(pool.StartSince), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.acceptedConnections, prometheus.CounterValue, float64(pool.AcceptedConnections), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.listenQueue, prometheus.GaugeValue, float64(pool.ListenQueue), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.maxListenQueue, prometheus.CounterValue, float64(pool.MaxListenQueue), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.listenQueueLength, prometheus.GaugeValue, float64(pool.ListenQueueLength), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.idleProcesses, prometheus.GaugeValue, float64(idle), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.activeProcesses, prometheus.GaugeValue, float64(active), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.totalProcesses, prometheus.GaugeValue, float64(total), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.maxActiveProcesses, prometheus.CounterValue, float64(pool.MaxActiveProcesses), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.maxChildrenReached, prometheus.CounterValue, float64(pool.MaxChildrenReached), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		ch <- prometheus.MustNewConstMetric(e.slowRequests, prometheus.CounterValue, float64(pool.SlowRequests), pool.Name, pool.ScrapeHost+pool.ScrapePath)
		// Opcache
		status := pool.CacheStatus.OPcacheStatus
		ch <- prometheus.MustNewConstMetric(enabledDesc, prometheus.GaugeValue, boolMetric(status.OPcacheEnabled))
		ch <- prometheus.MustNewConstMetric(cacheFullDesc, prometheus.GaugeValue, boolMetric(status.CacheFull))
		ch <- prometheus.MustNewConstMetric(restartPendingDesc, prometheus.GaugeValue, boolMetric(status.RestartPending))
		ch <- prometheus.MustNewConstMetric(restartInProgressDesc, prometheus.GaugeValue, boolMetric(status.RestartInProgress))
		ch <- prometheus.MustNewConstMetric(memoryUsageUsedMemoryDesc, prometheus.GaugeValue, intMetric(status.MemoryUsage.UsedMemory))
		ch <- prometheus.MustNewConstMetric(memoryUsageFreeMemoryDesc, prometheus.GaugeValue, intMetric(status.MemoryUsage.FreeMemory))
		ch <- prometheus.MustNewConstMetric(memoryUsageWastedMemoryDesc, prometheus.GaugeValue, intMetric(status.MemoryUsage.WastedMemory))
		ch <- prometheus.MustNewConstMetric(memoryUsageCurrentWastedPercentageDesc, prometheus.GaugeValue, status.MemoryUsage.CurrentWastedPercentage)
		ch <- prometheus.MustNewConstMetric(internedStringsUsageBufferSizeDesc, prometheus.GaugeValue, intMetric(status.InternedStringsUsage.BufferSize))
		ch <- prometheus.MustNewConstMetric(internedStringsUsageUsedMemoryDesc, prometheus.GaugeValue, intMetric(status.InternedStringsUsage.UsedMemory))
		ch <- prometheus.MustNewConstMetric(internedStringsUsageUsedFreeMemory, prometheus.GaugeValue, intMetric(status.InternedStringsUsage.FreeMemory))
		ch <- prometheus.MustNewConstMetric(statisticsNumCachedScripts, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.NumCachedScripts))
		ch <- prometheus.MustNewConstMetric(statisticsNumCachedKeys, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.NumCachedKeys))
		ch <- prometheus.MustNewConstMetric(statisticsMaxCachedKeys, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.MaxCachedKeys))
		ch <- prometheus.MustNewConstMetric(statisticsHits, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.Hits))
		ch <- prometheus.MustNewConstMetric(statisticsStartTime, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.StartTime))
		ch <- prometheus.MustNewConstMetric(statisticsLastRestartTime, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.LastRestartTime))
		ch <- prometheus.MustNewConstMetric(statisticsOOMRestarts, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.OOMRestarts))
		ch <- prometheus.MustNewConstMetric(statisticsHashRestarts, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.HashRestarts))
		ch <- prometheus.MustNewConstMetric(statisticsManualRestarts, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.ManualRestarts))
		ch <- prometheus.MustNewConstMetric(statisticsMisses, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.Misses))
		ch <- prometheus.MustNewConstMetric(statisticsBlacklistMisses, prometheus.GaugeValue, intMetric(status.OPcacheStatistics.BlacklistMisses))
		ch <- prometheus.MustNewConstMetric(statisticsBlacklistMissRatio, prometheus.GaugeValue, status.OPcacheStatistics.BlacklistMissRatio)
		ch <- prometheus.MustNewConstMetric(statisticsHitRate, prometheus.GaugeValue, status.OPcacheStatistics.OPcacheHitRate)

		for _, process := range pool.Processes {
			pidHash := calculateProcessHash(process)
			ch <- prometheus.MustNewConstMetric(e.processState, prometheus.GaugeValue, 1, pool.Name, pidHash, process.State, pool.ScrapeHost+pool.ScrapePath)
			ch <- prometheus.MustNewConstMetric(e.processRequests, prometheus.CounterValue, float64(process.Requests), pool.Name, pidHash, pool.ScrapeHost+pool.ScrapePath)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestMemory, prometheus.GaugeValue, float64(process.LastRequestMemory), pool.Name, pidHash, pool.ScrapeHost+pool.ScrapePath)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestCPU, prometheus.GaugeValue, process.LastRequestCPU, pool.Name, pidHash, pool.ScrapeHost+pool.ScrapePath)
			ch <- prometheus.MustNewConstMetric(e.processRequestDuration, prometheus.GaugeValue, float64(process.RequestDuration), pool.Name, pidHash, pool.ScrapeHost+pool.ScrapePath)
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
	// Opcache
	ch <- enabledDesc
	ch <- cacheFullDesc
	ch <- restartPendingDesc
	ch <- restartInProgressDesc
	ch <- memoryUsageUsedMemoryDesc
	ch <- memoryUsageFreeMemoryDesc
	ch <- memoryUsageWastedMemoryDesc
	ch <- memoryUsageCurrentWastedPercentageDesc
	ch <- internedStringsUsageBufferSizeDesc
	ch <- internedStringsUsageUsedMemoryDesc
	ch <- internedStringsUsageUsedFreeMemory
	ch <- internedStringsUsageUsedNumerOfStrings
	ch <- statisticsNumCachedScripts
	ch <- statisticsNumCachedKeys
	ch <- statisticsMaxCachedKeys
	ch <- statisticsHits
	ch <- statisticsStartTime
	ch <- statisticsLastRestartTime
	ch <- statisticsOOMRestarts
	ch <- statisticsHashRestarts
	ch <- statisticsManualRestarts
	ch <- statisticsMisses
	ch <- statisticsBlacklistMisses
	ch <- statisticsBlacklistMissRatio
	ch <- statisticsHitRate
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
