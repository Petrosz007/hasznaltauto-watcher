package main

import (
	"fmt"
	"regexp"
	"strings"

	colly "github.com/gocolly/colly/v2"
)

func CrawlSearchUrl(searchUrl string) []Listing {
	c := colly.NewCollector(
		colly.AllowedDomains("www.hasznaltauto.hu"),
		colly.Async(true),
	)

	listings := make([]Listing, 0, 10)

	c.OnHTML(`ul.pagination li:not(.first, .next, .prev) a[href]`, func(e *colly.HTMLElement) {
		c.Visit(e.Attr("href"))
	})

	yearRegex := regexp.MustCompile(`\d\d\d\d(\/\d?\d)?`)
	listingIdRegex := regexp.MustCompile(`(?P<ID>\d+)\#.*?$`)

	c.OnHTML("div.talalati-sor", func(e *colly.HTMLElement) {
		price := e.ChildText(`.price-fields-desktop .pricefield-primary`)
		if price == "" {
			price = e.ChildText(`.price-fields-desktop .pricefield-primary-highlighted`)
		}
		url := e.ChildAttr(`h3 > a`, "href")

		listing := Listing{
			ListingId:    listingIdRegex.FindStringSubmatch(url)[1],
			Title:        e.ChildText(`h3 > a`),
			Url:          url,
			ThumbnailUrl: e.ChildAttr(`.img__container img`, "src"),
			Price:        price,
		}

		e.ForEach(`.talalatisor-info.adatok > span`, func(i int, e *colly.HTMLElement) {
			text := strings.TrimSuffix(e.Text, ",")

			if i == 0 {
				listing.FuelType = text
			} else if strings.HasSuffix(text, "cm³") {
				listing.CubicCapacity = text
			} else if strings.HasSuffix(e.ChildText(`abbr`), "km") {
				listing.Milage = e.ChildText(`abbr`)
			} else if strings.HasSuffix(text, "kW") {
				listing.PowerkW = text
			} else if strings.HasSuffix(text, "LE") {
				listing.PowerHP = text
			} else if yearRegex.MatchString(text) {
				listing.Year = text
			}
		})

		e.ForEach(`span.label.label-hasznaltauto`, func(_ int, e *colly.HTMLElement) {
			if e.Text == "KIEMELT" {
				listing.IsHighlighted = true
			}
		})

		listings = append(listings, listing)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.Visit(searchUrl)

	c.Wait()

	return listings
}
