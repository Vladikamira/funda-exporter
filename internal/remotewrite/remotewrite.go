package remotewrite

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/castai/promwrite"
	"github.com/vladikamira/funda-exporter/internal/config"
)

func Send(results *[]config.House) {

	// send it over to prometheus
	client := promwrite.NewClient(*config.RemoteWriteUrl)

	for _, s := range *results {

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
							Name:  "post_code",
							Value: s.PostCode,
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
}
