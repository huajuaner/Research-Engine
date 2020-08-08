package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
)

type Node struct {
	IfValidate bool
	title string
	author string
	publisher string
	abstract string
}
type typeSafeSaver struct {
	mmp map [string] Node
	mux sync.Mutex
}

var (
	WebLimit = make (chan bool , 100 )
	SafeSaver =typeSafeSaver{mmp:make (map[string] Node)}
)

func UrlValidate (uri string )(Is bool ) {
	if uri != "" {
		Is ,_ = regexp.MatchString(`^(https?)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`,uri)
		if Is {
			Is ,_ = regexp.MatchString(`hitsz.edu.cn`,uri)
		}
		return
	}
	Is = false
	return
}
// https://book.douban.com/subject/3912973/
func BookValidate1(uri string) (Is bool) {
	if uri != "" {
		Is, _ = regexp.MatchString(`http://www.hitsz.edu.cn/article/view/id-[\d]+.html$`, uri)
		return
	}
	Is = false
	return
}



func MyCrawler(dst string, depth int, wg *sync.WaitGroup) {

	WebLimit <- true
	defer func(){
		<-WebLimit
	}()
	defer wg.Done()

	SafeSaver.mux.Lock()
	if _,ok := SafeSaver.mmp[dst];ok {
		defer SafeSaver.mux.Unlock()
		return
	}
	SafeSaver.mmp[dst] = Node{
		IfValidate: true,
	}
	SafeSaver.mux.Unlock()

	client := &http.Client{}

	req, err := http.NewRequest("GET", dst, nil)
	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println(dst)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		uri, If_exist := sel.Attr("href")
		if UrlValidate(uri) == false {
			NewUrl,err := url.Parse(uri)
			if err != nil {
				return
			}
			BaseUrl,err := url.Parse(dst)
			if err != nil {
				return
			}
			uri = BaseUrl.ResolveReference(NewUrl).String()
		}
		if UrlValidate(uri) == false {
			return
		}
		if If_exist == true && depth <= 2 {
			wg.Add(1)
			go MyCrawler(uri,depth+1,wg)
		}
	})
	if BookValidate1(dst) == false {
		return
	}
	// Find the review items
	doc.Find("div.detail_out").Each(func(i int, s *goquery.Selection) {
		title := s.Find("div.title").Text()
		fmt.Println(dst)
		fmt.Println(title)
	})
	return
}

func main() {
	var wg sync.WaitGroup
	wg.Add( 1 )
	go MyCrawler("http://www.hitsz.edu.cn/article/index.html",0,&wg)
	//go MyCrawler("http://portal.hitsz.edu.cn/portal",0,&wg)
	wg.Wait()
	fmt.Println("correct ending ")
}
