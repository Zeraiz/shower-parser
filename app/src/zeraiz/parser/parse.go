package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
	"sync"
)

type product struct {
	url          string
	name         string
	images       []string
	serial       string
	form         string
	sizes        string
	window       string
	bathType     string
	manufacturer string
	additions    []string
	description  string
}

var (
	w = sync.WaitGroup{}
)

func getUrlsForParse() []string {
	const mainUrl = "http://deto-shower.ru"
	doc := getHtml(mainUrl)
	aHrefs := doc.Find(".items.fix .actions a")
	result := make([]string, aHrefs.Length())
	aHrefs.Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		fullHref := strings.Join([]string{mainUrl, "/", href}, "")
		result[i] = fullHref
	})
	return result
}

func getHtml(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func parseData(url string, ch chan<- product) {
	doc := getHtml(url)

	ch <- product{
		url: url,
	}
}

func readData(ch <-chan product) {
	fmt.Println(<-ch)
}

func ParseDataFromUrls() {
	urls := getUrlsForParse()
	ch := make(chan product)
	fmt.Println("start")
	limitRoutine := make(chan int, 10)
	for _, url := range urls {
		url := url
		w.Add(1)

		go func() {
			limitRoutine <- 1
			readData(ch)
			<-limitRoutine
		}()
		go func() {
			defer w.Done()
			limitRoutine <- 1
			parseData(url, ch)
			<-limitRoutine
		}()
	}
	w.Wait()
	//for elem := range ch {
	//fmt.Println(elem)
	//}
}
