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

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for i := range in {
		wg.Add(1)
		go singleHashJob(i, out, wg, mu)
	}
	wg.Wait()
}

func singleHashJob(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	data := strconv.Itoa(in.(int))

	mu.Lock()
	md5Data := DataSignerMd5(data)
	mu.Unlock()

	crc32Chan := make(chan string)
	go crc32Worker(data, crc32Chan)

	crc32Data := <-crc32Chan
	crc32Md5Data := DataSignerCrc32(md5Data)

	out <- crc32Data + "~" + crc32Md5Data
}

func crc32Worker(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func MultiHash(in, out chan interface{}) {
	const TH int = 6
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)
		go multiHashJob(i.(string), out, TH, wg)
	}
	wg.Wait()
}

func multiHashJob(in string, out chan interface{}, th int, wg *sync.WaitGroup) {
	defer wg.Done()

	mu := &sync.Mutex{}
	jobWg := &sync.WaitGroup{}
	combinedChunks := make([]string, th)

	for i := 0; i < th; i++ {
		jobWg.Add(1)
		data := strconv.Itoa(i) + in

		go func(acc []string, index int, data string, jobWg *sync.WaitGroup, mu *sync.Mutex) {
			defer jobWg.Done()
			data = DataSignerCrc32(data)

			mu.Lock()
			acc[index] = data
			mu.Unlock()
		}(combinedChunks, i, data, jobWg, mu)
	}

	jobWg.Wait()
	out <- strings.Join(combinedChunks, "")
}

func CombineResults(in, out chan interface{}) {
	var slice []string
	for data := range in {
		slice = append(slice, data.(string))
	}
	sort.Strings(slice)
	result := strings.Join(slice, "_")
	fmt.Println(result)
	out <- result
}
