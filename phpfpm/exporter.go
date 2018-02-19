package phpfpm

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

const (
	namespace = "phpfpm"
)

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	PoolManager PoolManager
	mutex       sync.Mutex

	startSince          *prometheus.Desc
	acceptedConnections *prometheus.Desc
	listenQueue         *prometheus.Desc
	maxListenQueue      *prometheus.Desc
	listenQueueLength   *prometheus.Desc
	idleProcesses       *prometheus.Desc
	activeProcesses     *prometheus.Desc
	totalProcesses      *prometheus.Desc
	maxActiveProcesses  *prometheus.Desc
	maxChildrenReached  *prometheus.Desc
	slowRequests        *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,

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
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.PoolManager.Update()

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.startSince, prometheus.CounterValue, float64(pool.AcceptedConnections), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.acceptedConnections, prometheus.CounterValue, float64(pool.StartSince), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueue, prometheus.GaugeValue, float64(pool.ListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxListenQueue, prometheus.CounterValue, float64(pool.MaxListenQueue), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.listenQueueLength, prometheus.GaugeValue, float64(pool.ListenQueueLength), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.idleProcesses, prometheus.GaugeValue, float64(pool.IdleProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.activeProcesses, prometheus.GaugeValue, float64(pool.ActiveProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.totalProcesses, prometheus.GaugeValue, float64(pool.TotalProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxActiveProcesses, prometheus.CounterValue, float64(pool.MaxActiveProcesses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.maxChildrenReached, prometheus.CounterValue, float64(pool.MaxChildrenReached), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.slowRequests, prometheus.CounterValue, float64(pool.SlowRequests), pool.Name)
	}

	//if err := e.collect(ch); err != nil {
	//	log.Errorf("Error scraping apache: %s", err)
	//	e.scrapeFailures.Inc()
	//	e.scrapeFailures.Collect(ch)
	//}
	return
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
