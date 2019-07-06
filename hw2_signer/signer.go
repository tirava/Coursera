package main

import "sync"

//func main() {
//
//}

func Worker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	job(in, out)
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

}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}
