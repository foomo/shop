package persistence

import "github.com/prometheus/client_golang/prometheus"

const (
	labelDB         = "db"
	labelCollection = "collection"

	namespace        = "foomo_shop"
	subsystemService = "persistor"
)

var (
	getStandardSessionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemService,
			Name:      "get_standard_session_count",
			Help:      "count of standard session acquirements",
		}, []string{labelDB, labelCollection},
	)
	getGlobalSessionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemService,
			Name:      "get_global_session_count",
			Help:      "count of global session acquirements",
		}, []string{labelDB, labelCollection},
	)
)

func init() {
	prometheus.MustRegister(getStandardSessionCounter)
	prometheus.MustRegister(getGlobalSessionCounter)
}
