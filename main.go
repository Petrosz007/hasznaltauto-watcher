package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	colly "github.com/gocolly/colly/v2"
)

type Listing struct {
	ListingId     string
	Title         string
	Price         string
	ThumbnailUrl  string
	Url           string
	FuelType      string
	Year          string
	CubicCapacity string
	PowerkW       string
	PowerHP       string
	Milage        string
	SellerName    string
	IsHighlighted bool
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.hasznaltauto.hu"),
		// TODO: This caches a URL forever, so we won't get updates
		// ! Only use this for testing!
		colly.CacheDir("./colly-cache"),
	)

	listings := make([]Listing, 0, 10)

	c.OnHTML(`ul.pagination li.next a[href]`, func(e *colly.HTMLElement) {
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

	// testUrl := "https://www.hasznaltauto.hu/szemelyauto/toyota/prius/toyota_prius_1.5_hsd_ipa_navi_2006_automata-20441348#sid=5d6bebd7-940f-471b-998a-3300214fd26d"
	testSearchQueryUrl := "https://www.hasznaltauto.hu/talalatilista/PCOG2VG3R3RDADH5S56ADFGLZSCMOXNBIVNKDEKWTLL4UUCTIKJRQJLJGWAPR53V3LBDIFKPWWHOH6HY7CSCAZY76LTOKNETUKBIAJJZAVNWZREKWGURJ7UKA32QL2SAFU2JOATUPJYNXY2H4UFAZ2A3FDAF7RLKE4T72JINYMNFY3UURWLUZJAQ4MGHYDZTBMU3A5TZKK3X3FMIMXH75XGVQIHPSCM4VHNXXJEYW3PTOKKKOBQPOCVU6IXQ4ZDRKLIGIX7EXDTQ4HPKEMDHVEP6SCBPTEKXAFB4QH3NY2ZNAGSOAFHA7HAHHXYAILGUU2QB5GZ6WHA4AQ7QGEWE7Y6Q245PQSDAYZCCNE4WZL5XQCN7XQ57YKGHAL6SDA772A2WVDZ6KHXPG6DRMS7VJAJ5TWGDX33LD7EPYVJ6PQ5DEFAU3FG5DJMXNXK5OZW3NMYXS2FB7TNHKNOKSZV2KA52V2BMSGLQCXSKU6GQT264OKAKJACWVXLHTDIHIOOJIRLRHVFZ6YT5II7DMDBZXAEQFL2QIB2KKWVODTPI7YIO76OAKYHKCMTDG4TKTREOJSYOMZYA6YBOL4262ANM4FGEBLSEFXSU4725ME2KJ25TRC3R23C3GGGPXJ4B5Y3ZC4X4J2WOAGVXCRC3UKVV2GATJNZ6SQ7BDPMFOVHF5RPAGP2D2YND2JRIB7ASU2RUK6VCE2J6OVVQLDSP3NAIBYJC7EYHS75N3D3NKEC7Y365EQL5E5NWOFBVKYQXVRHHWDJZAFGFD25GMF23GWBAHGTT5SO5ZVR6YVUNC22F63JAY45MAT5PHCDIIOLI5ZFIDTRG7XNPQXP6EEHFG5IGSZMSKHR37AN3UG62BTYP4RFNAZWA6PPITTC675G7XWVADWIL37YH4H62H2Q"
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	// c.OnScraped(func(r *colly.Response) {
	// 	fmt.Println("Got Body", string(r.Body))
	// })
	c.Visit(testSearchQueryUrl)

	c.Wait()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(listings)
	// jsonEncodedListings, err := json.Marshal(listings)
	// if err != nil {
	// 	fmt.Println("Error: ", err.Error())
	// 	return
	// }
	// fmt.Printf("%v", string(jsonEncodedListings))
}
