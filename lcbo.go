package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Product struct {
	Name string
	Link string
	Price string
	Score string
}

func FetchProducts(category string, offset int) ([]*Product, error) {
	u, err := url.Parse("https://www.lcbo.com/webapp/wcs/stores/servlet/en/lcbo/" + category)
	if err != nil {
		return nil, err
	}
	var products []*Product
	query := url.Values{}
	query.Set("pageView", "grid")
	query.Set("facet_1", "vintagesonly%3A%22VINTAGES+ONLY%22")
	query.Set("facet_2", "instoreonly%3A%22IN+STORE+ONLY%22")
	query.Set("facetName_1", "VINTAGES+ONLY")
	query.Set("facetName_2", "IN+STORE+ONLY")
	query.Set("beginIndex", strconv.Itoa(offset))
	u.RawQuery = query.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", http.StatusText(res.StatusCode))
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	items := doc.Find("#content .productListingWidget .product_listing_container ul > li")
	items.Each(func(_ int, item *goquery.Selection) {
		chart := item.Find(".productChart")
		name := strings.TrimSpace(chart.Find(".product_name").Text())
		link, _ := chart.Find(".product_name > a").Attr("href")
		price := strings.TrimPrefix(strings.TrimSpace(chart.Find(".product_price > .price").Text()), "$")
		score := strings.TrimSpace(chart.Find(".product_score > .score").Text())
		products = append(products, &Product{
			Name: name,
			Link: link,
			Price: price,
			Score: score,
		})
	})
	return products, nil
}

func main() {

	w := csv.NewWriter(os.Stdout)
	if err := w.Write([]string{	"price", "score", "name", "link" }); err != nil {
		log.Fatal(err)
	}

	var offset int
	for {
		log.Print("Offset", offset)
		products, err := FetchProducts("wine-14", offset)
		if err != nil {
			log.Fatal(err)
		}
		if len(products) == 0 {
			break
		}
		for _, p := range products {
			if _, err := strconv.ParseFloat(p.Score, 64); err != nil {
				continue
			}
			if _, err := strconv.ParseFloat(p.Price, 64); err != nil {
				continue
			}
			if err := w.Write([]string{	p.Price, p.Score, p.Name, p.Link}); err != nil {
				log.Fatal(err)
			}
		}
		offset += len(products)
		w.Flush()
	}
}