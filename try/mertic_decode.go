package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

var (
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": "v0.1.0",
		},
	})

	alert = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alert",
		Help: "for alert purpose",
		ConstLabels: map[string]string{
			"reason": "test",
		},
	})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"code", "method"})

	testSummary = prometheus.NewSummary( prometheus.SummaryOpts{
		Name: "hello_world",
		ConstLabels: map[string]string{
			"reason": "test",
		},
	})
)

func main() {
	bind := ""
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&bind, "bind", ":8080", "The socket to bind to.")
	flagset.Parse(os.Args[1:])

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(version)
	r.MustRegister(alert)
	r.MustRegister(testSummary)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from example application."))
	})
	notfound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	setAlert := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alert set."))
		alert.Set(1)
	})

	unSetAlert := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alert unset."))
		alert.Set(0)
	})

	http.Handle("/", promhttp.InstrumentHandlerCounter(httpRequestsTotal, handler))
	http.Handle("/err", promhttp.InstrumentHandlerCounter(httpRequestsTotal, notfound))
	http.Handle("/alert/set", setAlert)
	http.Handle("/alert/unset", unSetAlert)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	// log.Fatal(http.ListenAndServe(bind, nil))

	mfs, err := r.Gather()
	if err != nil {
		log.Fatal(err)
	}

	for _, mf := range mfs {
		vec, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{
			model.Now(),
		}, mf)
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range vec {
			if s != nil {
				fmt.Println("metric:", s.Metric)
				fmt.Println("value:", s.Value)
				fmt.Println("timestamp:", s.Timestamp)
				fmt.Println("--------------------")
			}
		}
	}
}
