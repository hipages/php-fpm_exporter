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
	"github.com/speps/go-hashids"
)

const (
	namespace = "phpfpm"
)

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	mutex       sync.Mutex
	PoolManager PoolManager

	CalculateProcessScoreboard bool

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
	processState             *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,

		CalculateProcessScoreboard: false,

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
			"",
			[]string{"pool", "pid"},
			nil),

		processLastRequestMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_memory"),
			"",
			[]string{"pool", "pid"},
			nil),

		processLastRequestCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_last_request_cpu"),
			"",
			[]string{"pool", "pid"},
			nil),

		processState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_state"),
			"The process state.",
			[]string{"pool", "pid", "state"},
			nil),
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.PoolManager.Update()

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.scrapeFailues, prometheus.CounterValue, float64(pool.ScrapeFailures), pool.Name)

		if pool.ScrapeError != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, pool.Name)
			log.Errorf("Error scraping PHP-FPM: %v", pool.ScrapeError)
			continue
		}

		pps := CountProcessState(pool)
		if !e.CalculateProcessScoreboard && (pps.Active() != pool.ActiveProcesses || pps.Idle != pool.IdleProcesses) {
			log.Error("Inconsistent active and idle processes reported. Set `--fix-process-count` to have this calculated by php-fpm_exporter instead.")
		}

		var active, idle, total int64

		if e.CalculateProcessScoreboard {
			active = pool.ActiveProcesses
			idle = pool.IdleProcesses
			total = pool.TotalProcesses
		} else {
			active = pps.Active()
			idle = pps.Idle
			total = pps.Total()
		}

		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1, pool.Name)
		ch <- prometheus.MustNewConstMetric(e.startSince, prometheus.CounterValue, float64(pool.AcceptedConnections), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.acceptedConnections, prometheus.CounterValue, float64(pool.StartSince), pool.Name)
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
			pid := calculateProcessHash(process)
			ch <- prometheus.MustNewConstMetric(e.processState, prometheus.GaugeValue, 1, pool.Name, pid, process.State)
			ch <- prometheus.MustNewConstMetric(e.processRequests, prometheus.CounterValue, float64(process.Requests), pool.Name, pid)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestMemory, prometheus.GaugeValue, float64(process.LastRequestMemory), pool.Name, pid)
			ch <- prometheus.MustNewConstMetric(e.processLastRequestCPU, prometheus.GaugeValue, process.LastRequestCPU, pool.Name, pid)
		}
	}
}

// Describe exposes the metric description to Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
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
}

// calculateProcessHash generates a unique identifier for a process to ensure uniqueness across multiple systems/containers
func calculateProcessHash(pp PoolProcess) string {
	hd := hashids.NewData()
	hd.Salt = "php-fpm_exporter"
	hd.MinLength = 12
	h := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(pp.StartTime), int(pp.PID)})

	return e
}
