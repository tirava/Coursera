package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

type md5CallUnique struct {
	Mutex  sync.Mutex
	isBusy bool
}

type singleHash struct {
	Waiter        sync.WaitGroup
	Mutex         sync.Mutex
	cr32Result    string
	cr32Md5Result string
}

func (m *md5CallUnique) md5Calc(data string) (md5Result string) {
	m.Mutex.Lock()

	if !m.isBusy {
		m.isBusy = true

		md5Result = DataSignerMd5(data)

		fmt.Printf("%v SingleHash md5(data) %v\n", data, md5Result)
		m.isBusy = false
		m.Mutex.Unlock()

	}
	return

}

// Calculates crc32(data)
func (s *singleHash) cr32Calc(input string) {
	defer s.Waiter.Done()

	cr32Result := DataSignerCrc32(input)
	fmt.Printf("%v SingleHash crc32(data) %v\n", input, cr32Result)
	s.Mutex.Lock()
	s.cr32Result = cr32Result
	s.Mutex.Unlock()

}

//Calculates crc32(md5(data))
func (s *singleHash) Cr32Md5Calc(input string, m *md5CallUnique) {
	defer s.Waiter.Done()
	var md5Hash string

	isMd5Calculated := false
	for !isMd5Calculated {

		md5Hash = m.md5Calc(input)
		if md5Hash != "" {
			isMd5Calculated = true

		}
		runtime.Gosched()

	}
	cr32md5Hash := DataSignerCrc32(md5Hash)
	s.Mutex.Lock()
	s.cr32Md5Result = cr32md5Hash
	s.Mutex.Unlock()

}

// contats results of single hash calculations and put it to out chanel
func (s *singleHash) singleHashResultsContat(out chan interface{}, wg *sync.WaitGroup, inputData string) {

	s.Waiter.Wait()
	result := s.cr32Result + "~" + s.cr32Md5Result
	fmt.Printf("%v SingleHash result %v \n\n", inputData, result)

	out <- result
	wg.Done()

}

func singleHashCalc(data string, out chan interface{}, md5Uniq *md5CallUnique, wg *sync.WaitGroup) {

	defer wg.Done()

	singleHashCalculations := new(singleHash)

	fmt.Printf("%v SingleHash data %v \n", data, data)
	singleHashCalculations.Waiter.Add(1)
	go singleHashCalculations.cr32Calc(data)

	//Calculates crc32(md5(data))
	singleHashCalculations.Waiter.Add(1)
	go singleHashCalculations.Cr32Md5Calc(data, md5Uniq)

	waitForContat := &sync.WaitGroup{}
	waitForContat.Add(1)
	go singleHashCalculations.singleHashResultsContat(out, waitForContat, data)

	waitForContat.Wait()

}

// SingleHash is a job for calculation single hashes for each input
func SingleHash(in chan interface{}, out chan interface{}) {

	singleHashWG := &sync.WaitGroup{}

	md5CallObj := new(md5CallUnique)

	for dataRaw := range in {

		dataInt, ok := dataRaw.(int)
		if !ok {
			panic("SingleHash(): error at dataRaw.(int) - could not parse to int")
		}
		data := strconv.Itoa(dataInt)

		singleHashWG.Add(1)

		go singleHashCalc(data, out, md5CallObj, singleHashWG)
		runtime.Gosched()

	}
	singleHashWG.Wait()

}

type multiHash struct {
	sync.WaitGroup
	sync.Mutex
	multiHashes map[int]string
}

func (m *multiHash) multiHashcalc(input interface{}, out chan interface{}) {
	defer m.Done()
	singleHashData, ok := input.(string)
	if !ok {
		panic("main(): can't convert result data to string")
	}
	var multiHashRes string

	wg := &sync.WaitGroup{}
	m.Lock()
	m.multiHashes = make(map[int]string)
	m.Unlock()
	for th := 0; th <= 5; th++ {
		wg.Add(1)

		// Calculate cr32 for each step and put to map
		go m.cr32Calc(th, wg, singleHashData)

	}
	wg.Wait()
	var steps []int
	m.Lock()
	for step := range m.multiHashes {
		steps = append(steps, step)
	}
	m.Unlock()
	sort.Ints(steps)
	m.Lock()
	for _, step := range steps {

		multiHashRes += m.multiHashes[step]
	}
	m.Unlock()

	fmt.Printf("%s MultiHash result: %s \n\n", singleHashData, multiHashRes)

	out <- multiHashRes

}

func (m *multiHash) cr32Calc(step int, waiter *sync.WaitGroup, shData string) {
	defer waiter.Done()
	multiHashStep := DataSignerCrc32(strconv.Itoa(step) + shData)
	fmt.Printf("%v MultiHash: crc32(th+step1)) %v %v\n", shData, step, multiHashStep)
	m.Lock()
	m.multiHashes[step] = multiHashStep
	m.Unlock()

}

// MultiHash is a job for calculation multiHashes for each result of SingleHash for input
func MultiHash(in chan interface{}, out chan interface{}) {

	multiHashCalculations := new(multiHash)

	for dataRaw := range in {

		multiHashCalculations.Add(1)
		// Start multiHashes common goroutine

		go multiHashCalculations.multiHashcalc(dataRaw, out)

	}
	multiHashCalculations.Wait()

	fmt.Printf("\n")

}

// CombineResults is a job for combining results of all MultiHash calculations for all inputs
func CombineResults(in chan interface{}, out chan interface{}) {
	var multiHashesBeforeCombine []string

	var result string

	for dataRaw := range in {

		data, ok := dataRaw.(string)
		if !ok {
			panic("CombineResults: can't convert result data to string")
		}

		mu := &sync.Mutex{}
		mu.Lock()

		multiHashesBeforeCombine = append(multiHashesBeforeCombine, data)

		sort.Strings(multiHashesBeforeCombine)

		mu.Unlock()

	}

	for i, dataVal := range multiHashesBeforeCombine {
		if len(multiHashesBeforeCombine) > 0 {
			result += dataVal

		}
		if i != len(multiHashesBeforeCombine)-1 {
			result += "_"

		}

	}
	fmt.Printf("CombineResults %s \n\n", result)
	out <- result

}

// ExecutePipeline starts all jobs in pipeline that receive in array
func ExecutePipeline(jobs ...job) {

	waiter := &sync.WaitGroup{}

	in := make(chan interface{})
	out := make(chan interface{})

	for i, job := range jobs {
		waiter.Add(1)

		go startJob(job, waiter, in, out)

		in = out
		if i != len(jobs)-1 {
			out = make(chan interface{})

		}

	}

	waiter.Wait()

}

func startJob(job job, wg *sync.WaitGroup, in chan interface{}, out chan interface{}) {
	defer wg.Done()
	defer close(out)

	job(in, out)

}
