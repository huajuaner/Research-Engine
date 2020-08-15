package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/huichen/sego"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UrlNode struct {
	title     string
	dsturl    string
	frequency float64
	next      *UrlNode
}

type MyNode struct {
	Title     string
	Frequency float64
	Dsturl    string
}

var (
	inputReader = bufio.NewReader(os.Stdin)
	segmenter   sego.Segmenter
	JiebaMap    = make(map[string]*UrlNode)
)

func ReadMap(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("Failed To Open ", path)
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	record, err := reader.ReadAll()
	if err != nil {
		log.Println("Failed To Read ", path)
		panic(err)
	}

	for _, item := range record {
		frequency, _ := strconv.ParseFloat(item[2], 64)
		item[0] = strings.Replace(item[0], "\"", "", -1)
		node := UrlNode{
			title:     item[0],
			dsturl:    item[1],
			frequency: frequency,
			next:      JiebaMap[item[3]],
		}
		JiebaMap[item[3]] = &node
	}
}
func QueryResponse(query string) []MyNode {

	t1 := time.Now().UnixNano()
	segments := segmenter.Segment([]byte(query))
	slices := sego.SegmentsToSlice(segments, true)
	mmp := make(map[string]MyNode)
	for _, i := range slices {
		p := JiebaMap[i]
		for p != nil {
			cur, _ := mmp[p.dsturl]
			mmp[p.dsturl] = MyNode{
				Title:     p.title,
				Frequency: cur.Frequency + p.frequency,
			}
			p = p.next
		}
	}
	Base := make([]MyNode, 0)
	for i := range mmp {
		Base = append(Base, MyNode{
			Title:     mmp[i].Title,
			Frequency: mmp[i].Frequency,
			Dsturl:    i,
		})
	}

	sort.Slice(Base, func(i, j int) bool {
		return Base[i].Frequency > Base[j].Frequency
	})
	t2 := time.Now().UnixNano()
	fmt.Println(t1,"started ")
	fmt.Println(t2,"ended ")
	if len(Base) < 20 {
		return Base
	}

	return Base[0:19]
}

func main() {
	ReadMap("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/main/mmp")
	segmenter.LoadDictionary("C:/Users/zzy/Desktop/programs/GoResearchEngine/Research-Engine/src/crawler/dictionary.txt")

	r := gin.Default()
	r.LoadHTMLGlob("html/*")
	r.Static("imgs", "./imgs")

	r.GET("/index", func(c *gin.Context) {
		c.HTML(200, "index.html", "")
	})
	r.GET("/index/:query", func(c *gin.Context) {
		query := c.Param("query")
		Base := QueryResponse(query)
		c.HTML(200, "show.html", Base)
	})

	r.GET("/")
	r.Run(":8080")
}
