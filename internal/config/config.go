package config

import "flag"

type House struct {
	Price       int
	Address     string
	PostCode    string
	Link        string
	Area        int
	Year        int
	EnergyLabel string
}

var (
	FakeUserAgent = flag.String("fakeUserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"A fake User-Agent")
	FundaSearchUrl = flag.String("fundaSearchUrl", "https://www.funda.nl/koop/amstelveen/"+
		"200000-440000/70+woonopp/2+slaapkamers/",
		"Funda search page with paramethers")
	RemoteWriteUrl          = flag.String("remoteWriteUrl", "http://vmagent:8429/api/v1/write", "Url to send metrics via remoteWrite")
	ScrapeDelayMilliseconds = flag.Int("scrapeDelayMilliseconds", 1000, "Delay between scrapes. Let's not overload Funda :)")
)

// html objects from Funda
var (
	FundaHtmlSearchPages = ".search-output-result-count span"
)
