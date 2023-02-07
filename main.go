package main

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vladikamira/funda-exporter/internal/collector"
	"github.com/vladikamira/funda-exporter/internal/scraper"
)

var (
	FakeUserAgent = flag.String("fakeUserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36", "A fake User-Agent")
	FundaSearchUrl = flag.String("fundaSearchUrl", "https://www.funda.nl/koop/amstelveen,amsterdam/300000-440000/70+woonopp/2+slaapkamers/", "Funda search page with paramethers")
//	FundaSearchUrl = flag.String("fundaSearchUrl", "https://www.funda.nl/koop/amsterdam/buitenveldert-oost,buitenveldert-west,zuidas/300000-440000/70+woonopp/2+slaapkamers/", "Funda search page with paramethers")
	ScrapeDelayMilliseconds = flag.Int("scrapeDelayMilliseconds", 1000, "Delay between scrapes. Let's not overload Funda :)")
	ListenAddress = flag.String("listenAddress", ":2112", "Address to listen")
)

// main
func main() {

	var results []scraper.House

	// parse flags
	flag.Parse()

	// Setup better logging
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}

	log.SetFormatter(formatter)

	// Create a new instance of the fundaCollector and
	// register it with the prometheus client.
	fundaCollector := collector.NewFundaCollector(&results, FakeUserAgent, FundaSearchUrl, ScrapeDelayMilliseconds)
	prometheus.MustRegister(fundaCollector)

	// This section will start the HTTP server and expose
	// any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Starting on port " + *ListenAddress)
	log.Fatal(http.ListenAndServe(*ListenAddress, nil))

}
