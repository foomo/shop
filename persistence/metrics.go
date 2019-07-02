package persistence

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace        = "foomo_shop"
	subsystemService = "persistor"
)

var (
	getCollectionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemService,
			Name:      "get_collection_total",
			Help:      "Counts the number of invocations to 'getCollection (global/standard)' per database and collection.",
		}, []string{"db", "collection", "scope"},
	)
)

func init() {
	prometheus.MustRegister(getCollectionCounter)
}
