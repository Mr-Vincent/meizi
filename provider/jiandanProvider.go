package provider

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type JandanProvider struct{}

func (j *JandanProvider) UrlProvider() (url string) {
	return "http://jandan.net/girl/MjAyMjAxMTMtOTk=#comments"
}

func (j *JandanProvider) PagePagination(doc *goquery.Document, f func(nextPage string)) {
	doc.Find("#comments > div:nth-child(4) > div > a.previous-comment-page").Each(func(i int, s *goquery.Selection) {
		src, e := s.Attr("href")
		fmt.Println(src, e)
		if e {
			next := "http:" + src
			log.Printf("next page url -> %s\n", next)
			// 将解析到的下一页地址写出去
			f(next)
		}
	})
}

func (j *JandanProvider) ImageParser(doc *goquery.Document, f func(image string)) {
	doc.Find("#comments > ol >li > div > div.row > div.text > p > img").Each(func(i int, s *goquery.Selection) {
		src, e := s.Attr("src")
		if e {
			img := "http:" + src
			log.Printf("parse element : image url -> %s\n", img)
			f(img)
		}
	})
}
