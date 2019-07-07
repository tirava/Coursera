package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//func main() {
//
//}

func Worker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	//defer close(out)
	job(in, out)
	close(out)
}

func ExecutePipeline(jobs ...job) {

	in := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}

	for _, job := range jobs {
		out := make(chan interface{}, MaxInputDataLen)
		wg.Add(1)
		go Worker(job, in, out, wg)
		in = out
	}

	wg.Wait()
}

func crc32Worker(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func SingleHash(in, out chan interface{}) {
	//ch := make(chan string)
	//ch5 := make(chan string)
	//i := len(in)
	for data := range in {
		out <- strconv.Itoa(data.(int))
		//d := strconv.Itoa(data.(int))
		//s := DataSignerCrc32(d) + "~" + DataSignerCrc32(DataSignerMd5(d))
		//go crc32Worker(d, ch)
		//go crc32Worker(DataSignerMd5(d), ch5)
		//s := <-ch + "~" + <-ch5
		//fmt.Println(data, s)
		//out <- s
	}
	//for j := 0; j < i; j++ {
	//	s := <-ch + "~" + <-ch5
	//	fmt.Println(s)
	//	out <- s
	//	//println("out:", s)
	//}
	//println("------------------------")
}

func MultiHash(in, out chan interface{}) {
	//mch := make(map[int]chan string, MaxInputDataLen)
	//ch := make(chan string)
	for data := range in {
		out <- data
		//s := ""
		//for i := 0; i < 6; i++ {
		//	mch[i] = make(chan string)
		//	go crc32Worker(strconv.Itoa(i)+data.(string), mch[i])
		//	//s += <-ch
		//	//s += DataSignerCrc32(strconv.Itoa(i) + data.(string))
		//}
		//for i := 0; i < 6; i++ {
		//	s += <-mch[i]
		//}
		//fmt.Println(data, s)
		//out <- s
		//println("out:", s)
	}
	//println("------------------------")
}

func CombineResults(in, out chan interface{}) {
	slice := make([]string, 0)
	for data := range in {
		d := data.(string)
		s := DataSignerCrc32(d) + "~" + DataSignerCrc32(DataSignerMd5(d))
		sm := ""
		for i := 0; i < 6; i++ {
			sm += DataSignerCrc32(strconv.Itoa(i) + s)
			println(sm)
		}
		slice = append(slice, sm)
	}
	sort.Strings(slice)
	result := strings.Join(slice, "_")
	fmt.Println(result)
	out <- result
}
