package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"regexp"
	"sync"
)

func UrlValidate (url string )(Is bool ) {
	if url != "" {
		Is ,_ = regexp.MatchString(`(https?)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`,url)
		return
	}
	Is = false
	return
}
// https://book.douban.com/subject/3912973/
func BookValidate1(url string) (Is bool) {
	if url != "" {
		Is, _ = regexp.MatchString(`https://book.douban.com/subject/[\d]+/$`, url)
		return
	}
	Is = false
	return
}

var WebLimit = make (chan bool , 10 )


func MyCrawler(dst string, depth int, wg *sync.WaitGroup) {

	WebLimit <- true
	defer func(){
		<-WebLimit
	}()
	defer wg.Done()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", dst, nil)
	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")
	res, err := client.Do(req)
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

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		url, If_exist := sel.Attr("href")
		if UrlValidate(url) == false {
			return
		}
		if If_exist == true && depth <= 2 {
			wg.Add(1)
			fmt.Println(url)
			go MyCrawler(url,depth+1,wg)
		}
		//if If_exist == true && BookValidate1(url) {
		//	fmt.Println(url)
		//}
	})
	if BookValidate1(dst) == false {
		return
	}
	// Find the review items
	//doc.Find("div.info").Each(func(i int, s *goquery.Selection) {
	//	// For each item found, get the band and title
	//	url, IfExist := s.Find("div.title a").Attr("href")
	//	if IfExist == true {
	//		title := s.Find("a").Text()
	//		author := s.Find("div.author").Text()
	//		fmt.Println(i, title, author, url)
	//	}
	//})
	return
}

func main() {
	var wg sync.WaitGroup
	wg.Add( 1 )
	go MyCrawler("https://book.douban.com/top250?icn=index-book250-all",0,&wg )
	wg.Wait()
}
