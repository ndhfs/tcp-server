package tcp

import "github.com/prometheus/client_golang/prometheus"

type metricsManger struct {
	prom               prometheus.Registerer
	acceptCount        prometheus.Counter
	acceptDeclineCount prometheus.Counter
	workerCount        prometheus.Gauge
	workerIdleCount    prometheus.Gauge
	connectionsCount   prometheus.Gauge
}

func (m *metricsManger) init(o Options) {
	m.prom = o.prom
	if m.prom == nil {
		return
	}

	m.acceptCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "tcp_server_" + o.name,
		Name: "accept_count",
	})

	m.acceptDeclineCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "tcp_server_" + o.name,
		Name: "accept_decline_count",
	})

	m.workerCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "tcp_server_" + o.name,
		Name: "workers_count",
	})
	m.workerCount.Set(float64(o.workersNum))

	m.workerIdleCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "tcp_server_" + o.name,
		Name: "workers_idle_count",
	})
	m.workerIdleCount.Set(float64(o.workersNum))

	m.connectionsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "tcp_server_" + o.name,
		Name: "connections_count",
	})
	m.workerIdleCount.Set(float64(o.workersNum))

	m.prom.MustRegister(
		m.acceptCount,
		m.acceptDeclineCount,
		m.workerCount,
		m.workerIdleCount,
		m.connectionsCount,
	)
}

func (m *metricsManger) acceptInc() {
	if m.prom == nil {
		return
	}
	m.acceptCount.Inc()
}

func (m *metricsManger) acceptDeclineInc() {
	if m.prom == nil {
		return
	}
	m.acceptDeclineCount.Inc()
}

func (m *metricsManger) workersIdleInc() {
	if m.prom == nil {
		return
	}
	m.workerIdleCount.Inc()
}

func (m *metricsManger) workersIdleDec() {
	if m.prom == nil {
		return
	}
	m.workerIdleCount.Dec()
}

func (m *metricsManger) connectionsInc() {
	if m.prom == nil {
		return
	}
	m.connectionsCount.Inc()
}

func (m *metricsManger) connectionsDec() {
	if m.prom == nil {
		return
	}
	m.connectionsCount.Dec()
}
