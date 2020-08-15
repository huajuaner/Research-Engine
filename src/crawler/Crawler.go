package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/huichen/sego"
)

type Node struct {
	IfValidate bool
	title string
	author string
	publisher string
	abstract string
}

type UrlNode struct {
	title string
	dsturl string
	frequency float64
	next  * UrlNode
}

type typeSafeSaver struct {
	mmp map [string] Node
	mux sync.Mutex
}
type SeJiebamap struct {
	mmp map [string] *UrlNode
	mux sync.Mutex
}

var (
	segmenter sego.Segmenter
	WebLimit = make (chan bool , 100 )
	SafeSaver =typeSafeSaver{mmp:make (map[string] Node)}
	Jiebamap = SeJiebamap{mmp:make(map[string] *UrlNode)}
	SumOfPages int
)

func UrlValidate (uri string )(Is bool ) {
	if uri != "" {
		Is ,_ = regexp.MatchString(`^(https?)://[-A-Za-z-1-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`,uri)
		if Is {
			Is ,_ = regexp.MatchString(`hitsz.edu.cn`,uri)
		}
		return
	}
	Is = false
	return
}

func FurtherValidate(uri string) (Is bool) {
	if uri != "" {
		Is, _ = regexp.MatchString(`http://www.hitsz.edu.cn/article/view/id-[\d]+.html$`, uri)
		return
	}
	Is = false
	return
}

func SquuzeSegments(dst , title string,base [] string ){
	mmp := make ( map[string] float64)
	for _,i := range base{
		mmp[i] += 1
	}
	titleSeg := segmenter.Segment([]byte(title))
	titlestring := sego.SegmentsToSlice(titleSeg,true)

	for _,i := range titlestring {
		mmp[i] += 3
	}
	length := float64( len(base) )
	for i := range mmp {
		Jiebamap.mux.Lock()
		j,ok := Jiebamap.mmp[i]
		if ok == false {
			j = nil
		}
		cur := UrlNode{
			title:     title,
			dsturl:    dst,
			frequency: mmp[i]/length,
			next:      j,
		}
		Jiebamap.mmp[i] = &cur
		Jiebamap.mux.Unlock()
	}
}

func MyCrawler(dst string, depth int, wg *sync.WaitGroup) {
	defer fmt.Println(dst ," crawler finished ")
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
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}


	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println(dst ," response was not 200")
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		uri, IfExist := sel.Attr("href")
		if IfExist == false {
			return
		}
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
		if IfExist == true && depth <= 1 {
			wg.Add(1)
			go MyCrawler(uri,depth+1,wg)
		}
	})
	if FurtherValidate(dst) == false {
		return
	}
	// Find the review items
	SumOfPages ++
	doc.Find("div.detail_out").Each(func(i int, s *goquery.Selection) {
		title := s.Find("div.title").Text()
		context := s.Find("div.edittext").Text()
		context = strings.Replace(context," ","",-1)
		context = strings.Replace(context,"\n","",-1)
		context = strings.Replace(context,"\r","",-1)
		segments:= segmenter.Segment([] byte(context) )
		slices := sego.SegmentsToSlice(segments,true)
		SquuzeSegments(dst,title,slices)
	})
	return
}

func WriteJBMap (path string ) {
	defer fmt.Println(path," writen coorrectly")
	csvFile,err := os.Create(path)
	if err != nil{
		panic (err)
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	for i:= range Jiebamap.mmp{
		p := Jiebamap.mmp[i]
		for p!=nil {
			line := [] string{"\""+p.title+"\"",p.dsturl,strconv.FormatFloat(p.frequency,'E',-1,64), i}
			err := writer.Write(line)
			if err !=nil {
				panic (err)
			}
			p = p.next
		}
	}
	writer.Flush()
}

func Organize () {
	for i := range Jiebamap.mmp{
		count := 0
		p := Jiebamap.mmp[i]
		for p != nil {
			count ++
			p = p.next
		}
		bias := math.Log(float64(SumOfPages)/float64(count+1))
		p = Jiebamap.mmp[i]
		for p!= nil {
			p.frequency *= bias
			p = p.next
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add( 1 )
	segmenter.LoadDictionary("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/crawler/dictionary.txt")
	//go MyCrawler("http://www.hitsz.edu.cn/article/view/id-98581.html",0,&wg)
	go MyCrawler("http://www.hitsz.edu.cn/article/index.html",0,&wg)
	//go MyCrawler("http://portal.hitsz.edu.cn/portal",-1,&wg)
	wg.Wait()
	fmt.Println("correct ending ")
	Organize()
	WriteJBMap("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/main/mmp")
}