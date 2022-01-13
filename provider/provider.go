package provider

import "github.com/PuerkitoBio/goquery"

type Provider interface {
	UrlProvider() (url string)
	PagePagination(doc *goquery.Document, f func(nextPage string))
	ImageParser(doc *goquery.Document, f func(image string))
}
