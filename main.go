package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/atotto/encoding/csv"

	"github.com/vladikamira/funda-exporter/internal/config"
	"github.com/vladikamira/funda-exporter/internal/remotewrite"
	"github.com/vladikamira/funda-exporter/internal/scraper"
)

// main
func main() {

	var results []config.House

	// parse flags
	flag.Parse()

	res, err := scraper.ScrapePageContent(*config.FundaSearchUrl, *config.FakeUserAgent)
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
	pages, _ := strconv.Atoi(numberRegex.FindString(doc.Find(config.FundaHtmlSearchPages).Text()))
	resultsOnPage := 15

	cicles := 0
	if pages%resultsOnPage == 0 {
		cicles = (pages / resultsOnPage)
	} else {
		cicles = (pages / resultsOnPage) + 1
	}

	fmt.Printf("Found %v results on %v pages\n", pages, cicles)

	for i := 1; i <= cicles; i++ {
		scraper.ScrapeFunda(fmt.Sprintf(*config.FundaSearchUrl+"p%d/", i), &results)
	}

	// save result in file
	f, _ := os.Create("house.txt")
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteStructAll(results)

	//	fmt.Println(results)
	remotewrite.Send(&results)

}
