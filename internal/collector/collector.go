package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vladikamira/funda-exporter/internal/scraper"
)

type FundaCollector struct {
	results           *[]scraper.House
	userAgent         *string
	searchUrl         *string
	delayMilliseconds *int
	fundaPrice        *prometheus.Desc
}

// FundaCollector You must create a constructor for your collector that
// initializes every descriptor and returns a pointer to the collector
func NewFundaCollector(r *[]scraper.House, userAgent, searchUrl *string, delay *int) *FundaCollector {

	return &FundaCollector{
		results:           r,
		userAgent:         userAgent,
		searchUrl:         searchUrl,
		delayMilliseconds: delay,
		fundaPrice: prometheus.NewDesc("funda_apartment_price",
			"Funda Apartment price",
			[]string{
				"address",
				"post_code",
				"link",
				"energy_label",
				"year",
				"area",
				"published",
				"isolation",
				"extra_payments",
				"city",
			}, nil,
		),
	}
}

// Describe Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *FundaCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the metric you create for a given collector
	ch <- collector.fundaPrice
}

// Collect implements required collect function for all prometheus collectors
func (collector *FundaCollector) Collect(ch chan<- prometheus.Metric) {

	scraper.RunScraper(collector.results, collector.userAgent, collector.searchUrl, collector.delayMilliseconds)

	for _, s := range *collector.results {

		// Write the latest value for each metric in the prometheus metric channel.
		// Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
		ch <- prometheus.MustNewConstMetric(
			collector.fundaPrice,
			prometheus.GaugeValue,
			float64(s.Price),
			s.Address,
			s.PostCode,
			s.Link,
			s.EnergyLabel,
			strconv.Itoa(s.Year),
			strconv.Itoa(s.Area),
			s.Published,
			s.Isolation,
			s.ExtraPayments,
			s.City,
		)
	}
}
