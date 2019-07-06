package main

import "sync"

//func main() {
//
//}

//func Worker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
//	defer wg.Done()
//	println("start Worker", in)
//	job(in, out)
//	println("end Worker", in)
//	//close(in)
//}

func Worker(joba chan job, in, out chan interface{}, wg *sync.WaitGroup) {
	//defer wg.Done()
	//in := make(chan interface{}, MaxInputDataLen)
	//out := make(chan interface{}, MaxInputDataLen)
	println("bef joba")
	for job := range joba {
		println("worker joba")
		job(in, out)
		//close(out)
	}
	println("exit worker")
}

func ExecutePipeline(jobs ...job) {

	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}

	workerInput := make(chan job, len(jobs))
	//for range jobs {
	//	wg.Add(1)
	go Worker(workerInput, in, out, wg)
	//}

	for _, job := range jobs {
		//println("bef input")
		//in, out = out, in
		//wg.Add(1)
		//go Worker(job, in, out, wg)
		//go Worker(job)
		//runtime.Gosched()
		workerInput <- job
		//in, out = out, in
		//println("after input")
	}

	close(workerInput)

	//wg.Wait()
}

func SingleHash(in, out chan interface{}) {

}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}
