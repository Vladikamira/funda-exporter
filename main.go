package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	//	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/atotto/encoding/csv"
	"github.com/castai/promwrite"
)

type House struct {
	Price       int
	Address     string
	Link        string
	Area        int
	Year        int
	EnergyLabel string
}

var (
	fakeUserAgent = flag.String("fakeUserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"A fake User-Agent")
	fundaSearchUrl = flag.String("fundaSearchUrl", "https://www.funda.nl/koop/amstelveen/"+
		"200000-440000/70+woonopp/2+slaapkamers/",
		"Funda search page with paramethers")
	remoteWriteUrl          = flag.String("remoteWriteUrl", "http://vmagent:8429/api/v1/write", "Url to send metrics via remoteWrite")
	scrapeDelayMilliseconds = flag.Int("scrapeDelayMilliseconds", 1000, "Delay between scrapes. Let's not overload Funda :)")
)

// just make a request and
func ScrapePageContent(url string) (*http.Response, error) {

	fmt.Printf("Scraping %s\n", url)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	req.Header.Set("User-Agent", *fakeUserAgent)

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	//	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, nil
	}

	return res, nil
}

// parsing search page
func ScrapeFunda(url string, result *[]House) {

	res, err := ScrapePageContent(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// prepare regular expressions
	numberRegex, _ := regexp.Compile("[0-9\\.]+")
	notNumberRegex, _ := regexp.Compile("[^0-9]")
	space, _ := regexp.Compile(`\s+`)

	// do parsing search page
	doc.Find(".search-result").Each(func(i int, s *goquery.Selection) {

		var h House
		h.Address = space.ReplaceAllString(s.Find(".search-result__header-title-col").Text(), " ")
		h.Link, _ = s.Find(".search-result__header a").Attr("href")
		h.Link = "https://www.funda.nl" + h.Link

		firstPrice := numberRegex.FindString(s.Find(".search-result-price").Text())
		priceString := notNumberRegex.ReplaceAllString(firstPrice, "")
		h.Price, _ = strconv.Atoi(priceString)

		GetHouseDetail(&h)

		*result = append(*result, h)
	})
}

func GetHouseDetail(h *House) {
	url := h.Link

	res, err := ScrapePageContent(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	yearRegex, _ := regexp.Compile("[0-9]{4}")
	numberRegex, _ := regexp.Compile("[0-9]+")
	energyLabelRegex, _ := regexp.Compile("[A-G]{1}[+]*")
	space, _ := regexp.Compile(`\s+`)

	doc.Find(".object-kenmerken-list dt").Each(func(i int, s *goquery.Selection) {
		nf := s.NextFiltered("dd")

		key := space.ReplaceAllString(s.Text(), " ")
		value := space.ReplaceAllString(nf.Text(), " ")

		switch key {
		case "Wonen": // square meters
			h.Area, _ = strconv.Atoi(numberRegex.FindString(value))
		case "Energielabel": // energy label
			h.EnergyLabel = energyLabelRegex.FindString(value)
		case "Bouwjaar": // costruction year
			h.Year, _ = strconv.Atoi(yearRegex.FindString(value))
		default:
			//
		}

		// debug
		//fmt.Println("KEY: ", key, "VALUE: ", value)

	})
	//fmt.Println("   ")
	time.Sleep(time.Duration(*scrapeDelayMilliseconds) * time.Millisecond)
}

func main() {

	var results []House

	// parse flags
	flag.Parse()

	res, err := ScrapePageContent(*fundaSearchUrl)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// find amount of elements in search
	numberRegex, _ := regexp.Compile("[0-9]+")
	pages, _ := strconv.Atoi(numberRegex.FindString(doc.Find(".search-output-result-count span").Text()))
	resultsOnPage := 15
	cicles := (pages / resultsOnPage) + 1
	fmt.Printf("Found %v results on %v pages\n", pages, cicles)

	for i := 1; i <= cicles; i++ {
		ScrapeFunda(fmt.Sprintf(*fundaSearchUrl+"p%d/", i), &results)
	}

	f, _ := os.Create("house.txt")
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteStructAll(results)

	// send it over to prometheus
	client := promwrite.NewClient(*remoteWriteUrl)

	for _, s := range results {
		//		fmt.Println(s)
		_, err := client.Write(context.Background(), &promwrite.WriteRequest{
			TimeSeries: []promwrite.TimeSeries{
				{
					Labels: []promwrite.Label{
						{
							Name:  "__name__",
							Value: "funda_apartment_price",
						},
						{
							Name:  "address",
							Value: s.Address,
						},
						{
							Name:  "link",
							Value: s.Link,
						},
						{
							Name:  "energy_label",
							Value: s.EnergyLabel,
						},
						{
							Name:  "year",
							Value: strconv.Itoa(s.Year),
						},
						{
							Name:  "area",
							Value: strconv.Itoa(s.Area),
						},
					},
					Sample: promwrite.Sample{
						Time:  time.Now(),
						Value: float64(s.Price),
					},
				},
			},
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	//	fmt.Println(results)

}
