package scraper

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vladikamira/funda-exporter/internal/config"
)

// just make a request and
func ScrapePageContent(url, fakeUserAgent string) (*http.Response, error) {

	fmt.Printf("Scraping %s\n", url)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	req.Header.Set("User-Agent", fakeUserAgent)

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
func ScrapeFunda(url string, result *[]config.House) {

	res, err := ScrapePageContent(url, *config.FakeUserAgent)
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
	postCodeRegex, _ := regexp.Compile("[0-9]{4} [A-Z]{2}")
	space, _ := regexp.Compile(`\s+`)

	// do parsing search page
	doc.Find(".search-result").Each(func(i int, s *goquery.Selection) {

		var h config.House
		h.Address = space.ReplaceAllString(s.Find(".search-result__header-title-col").Text(), " ")
		h.Link, _ = s.Find(".search-result__header a").Attr("href")
		h.Link = "https://www.funda.nl" + h.Link

		firstPrice := numberRegex.FindString(s.Find(".search-result-price").Text())
		priceString := notNumberRegex.ReplaceAllString(firstPrice, "")
		h.Price, _ = strconv.Atoi(priceString)

		h.PostCode = postCodeRegex.FindString(h.Address)
		//fmt.Println(h.PostCode)

		GetHouseDetail(&h)

		*result = append(*result, h)
	})
}

func GetHouseDetail(h *config.House) {
	url := h.Link

	res, err := ScrapePageContent(url, *config.FakeUserAgent)
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
	time.Sleep(time.Duration(*config.ScrapeDelayMilliseconds) * time.Millisecond)
}
