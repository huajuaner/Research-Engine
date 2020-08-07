package main

import (
	"fmt"
	"sync"
)

func myfunc( depth int , wg * sync.WaitGroup ) {
	if depth >=3 {
		return
	}
	defer wg.Done()
	fmt.Println(depth)
	if depth <2 {
		wg.Add(2)
	}
	go myfunc( depth+1 , wg)
	go myfunc( depth+1 , wg)
}
func main (){
	var wg sync.WaitGroup
	wg.Add(1)
	go myfunc( 0 , &wg)
	wg.Wait()
}