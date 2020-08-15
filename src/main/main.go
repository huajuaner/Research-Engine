package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/huichen/sego"
	"io"
	"log"
	"os"
	"sort"
)

type UrlNode struct {
	title     string
	dsturl    string
	frequency float64
	pos [] int
	next      *UrlNode
}

type MyNode struct {
	title string
	frequency float64
	dsturl string
}
var (
	inputReader = bufio.NewReader(os.Stdin)
	segmenter  sego.Segmenter
	JiebaMap = make(map[string]*UrlNode)
)

func ReadMap (path string ) {
	file,err := os.Open( path )
	if err != nil {
		log.Println("Failed To Open ",path)
		panic (err)
	}
	defer file .Close()

	chunks := make ([]byte , 0)
	buf := make ([]byte,1024)
	for {
		n,err := file.Read(buf)
		if err !=nil && err != io.EOF{
			log.Fatalln(err)
		}
		if 0 == n {
			break
		}
		chunks = append(chunks, buf[:n]...)
	}
	type CurNode struct {
		title string
		frequency float64
		dsturl string
		pos []int
	}
	base := make([]CurNode,0)
	err = json.Unmarshal(chunks,&base)
	if err !=nil {
		log.Fatalln(err)
	}
	fmt.Println(base)
	//reader := csv.NewReader(file)
	//reader.FieldsPerRecord = -1
	//record , err := reader.ReadAll()
	//if err != nil {
	//	log.Println("Failed To Read ",path)
	//	panic(err)
	//}
	//
	//for _,item := range record {
	//	frequency, _ := strconv.ParseFloat(item[2],64)
	//	item[0] = strings.Replace(item[0],"\"","",-1)
	//	node := UrlNode{
	//		title:     item[0],
	//		dsturl:    item[1],
	//		frequency: frequency,
	//		next:      JiebaMap[item[3]],
	//	}
	//	JiebaMap[item[3]] = & node
	//}
}

func PrintMap () {
	for i:= range JiebaMap {
		p := JiebaMap[i]
		for p!=nil {
			fmt.Println( *p,i )
			p = p.next
		}
	}
}
func QueryResponse (query string ){
	segments := segmenter.Segment([] byte(query))
	slices := sego.SegmentsToSlice(segments,true)
	fmt.Println(slices)
	mmp := make (map [string] MyNode )
	for _,i := range slices {
		p := JiebaMap[i]
		for p!=nil {
			cur,_ := mmp[p.dsturl]
			mmp[p.dsturl] = MyNode{
				title:     p.title,
				frequency: cur.frequency+p.frequency,
			}
			p = p.next
		}
	}
	Base := make ([]MyNode,0)
	for i := range mmp {
		Base = append(Base, MyNode{
			title:     mmp[i].title,
			frequency: mmp[i].frequency,
			dsturl:    i,
		})
	}
	sort.Slice(Base , func (i,j int)bool {
		return Base[i].frequency>Base[j].frequency
	})
	for _,i:= range Base {
		fmt.Println(i)
	}
}
func Response (){
	query,err := inputReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}
	QueryResponse(query)

}
func main() {
	ReadMap("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/main/backupmmp")
	segmenter.LoadDictionary("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/crawler/dictionary.txt")
	//PrintMap()
	Response()
}
