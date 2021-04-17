package parser

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
	"sync"
)

type product struct {
	url         string
	name        string
	images      []string
	serial      string
	shape       string
	sizes       string
	window      string
	description string
	price       int
}

const mainUrl = "http://deto-shower.ru"

var (
	w  sync.WaitGroup
	mu sync.Mutex
)

func getUrlsForParse() []string {
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

	mainOptions := doc.Find(".product .R")

	// Получение названия
	name := mainOptions.Find("h1").Text()

	// Получение формы
	shape := mainOptions.Find("#toc1").Text()[12:]

	// Получение размеров
	sizes := mainOptions.Find("#toc2").Text()[17:]
	sizes = strings.TrimRight(sizes, " см")
	// Получение типа окна
	window := mainOptions.Find("#toc3").Text()

	// Получение описания
	description, err := doc.Find(".text.basement").Html()
	if err != nil {
		log.Fatal(err)
	}

	// Получение картинок альбома
	docImages := doc.Find(".img.cycle a")
	images := make([]string, docImages.Length())
	docImages.Each(func(i int, s *goquery.Selection) {
		image, _ := s.Attr("href")
		images[i] = strings.Join([]string{mainUrl, "/", image}, "")
	})

	// Получение серии
	serial := doc.Find("#toc0").Text()

	price, _ := strconv.Atoi(doc.Find(".R .price span").Text())

	ch <- product{
		name:        name,
		url:         url,
		images:      images,
		serial:      serial,
		shape:       shape,
		sizes:       sizes,
		window:      window,
		description: description,
		price:       price,
	}
}

func buildFileName(fUrl string) string {
	fileUrl, err := url.Parse(fUrl)
	if err != nil {
		log.Fatal(err)
	}
	path := fileUrl.Path
	segments := strings.Split(path, "/")
	return segments[len(segments)-1]
}

func getSize(sizes []string, index int) string {
	length := len(sizes)

	if length == 0 || length < 2 {
		return ""
	}
	if index == 0 {
		return sizes[0]
	}

	if index == 1 && length > 1 {
		return sizes[1]
	}

	if index == 2 && length > 2 {
		return sizes[2]
	}

	return ""
}

func workData(ch <-chan product, writer *csv.Writer) {
	product := <-ch
	mu.Lock()
	defer mu.Unlock()
	//"model",
	//	"name",
	//	"image",
	//	"price",
	//	"weight",
	//	"height",
	//	"length",
	//	"attribute:Glass",
	//	"attribute:Type",
	//	"description",
	//
	//fmt.Println(product.shape)
	//return

	sizes := strings.Split(product.sizes, "х")

	additionalImage := ""
	if len(product.images) > 1 {
		additionalImage = product.images[1]
	}
	name := strings.TrimRight(product.name, " ")

	err := writer.Write([]string{
		getCell(name),
		getCell(name),
		getCell(product.images[0]),
		getCell(additionalImage),
		"1",
		"Душевые кабины",
		getCell(name),
		getCell(product.description),
		strconv.Itoa(product.price),
		getCell(getSize(sizes, 1)),
		getCell(getSize(sizes, 0)),
		getCell(getSize(sizes, 2)),
		getCell(strings.TrimLeft(product.window, "Стекла: ")),
	})
	if len(product.images) < 3 {
		return
	}
	for _, sUrl := range product.images[2:] {
		sUrl := sUrl
		err := writer.Write([]string{
			getCell(name),
			getCell(name),
			getCell(""),
			getCell(sUrl),
		})
		if err != nil {
			fmt.Println(err)
			log.Fatalln(err)
		}
	}
	if err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}
}

func getCell(val string) string {
	if val == "" {
		return " "
	}
	return val
}

func ParseDataFromUrls() {
	urls := getUrlsForParse()
	ch := make(chan product)
	fmt.Println("start")
	limitRoutineR := make(chan int, 10)
	limitRoutineW := make(chan int, 10)

	f, _ := os.Create("data/showers_data.csv")
	bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
	f.Write(bomUtf8)
	defer f.Close()
	csvWriter := csv.NewWriter(f)
	csvWriter.Comma = ';'

	defer csvWriter.Flush()
	_ = csvWriter.Write([]string{
		"model",
		"name",
		"image",
		"additional_image",
		"status",
		"category",
		"Meta Title",
		"description",
		"price",
		"width",
		"height",
		"length",
		"attribute:Glass",
	})

	for _, sUrl := range urls {
		sUrl := sUrl

		go func() {
			w.Add(1)
			defer w.Done()
			limitRoutineR <- 1
			workData(ch, csvWriter)
			<-limitRoutineR
		}()
		go func() {
			w.Add(1)
			defer w.Done()
			limitRoutineW <- 1
			parseData(sUrl, ch)
			<-limitRoutineW
		}()
	}
	w.Wait()
}
