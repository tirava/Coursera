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
	for data := range in {
		d := strconv.Itoa(data.(int))
		s := DataSignerCrc32(d) + "~" + DataSignerCrc32(DataSignerMd5(d))
		//go crc32Worker(d, ch)
		//s := DataSignerCrc32(<-ch) + "~" + DataSignerCrc32(DataSignerMd5(d))
		fmt.Println(data, s)
		out <- s
	}
}

func MultiHash(in, out chan interface{}) {
	mch := make(map[int]chan string, MaxInputDataLen)
	//ch := make(chan string)
	for data := range in {
		s := ""
		for i := 0; i < 6; i++ {
			mch[i] = make(chan string)
			go crc32Worker(strconv.Itoa(i)+data.(string), mch[i])
			//s += <-ch
			//s += DataSignerCrc32(strconv.Itoa(i) + data.(string))
		}
		for i := 0; i < 6; i++ {
			s += <-mch[i]
		}
		fmt.Println(data, s)
		out <- s
	}
}

func CombineResults(in, out chan interface{}) {
	slice := make([]string, 0)
	for data := range in {
		slice = append(slice, data.(string))
	}
	sort.Strings(slice)
	result := strings.Join(slice, "_")
	fmt.Println(result)
	out <- result
}
