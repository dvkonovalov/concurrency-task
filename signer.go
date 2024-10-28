package concurrency_task

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ExecutePipeline
func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup

	in := make(chan interface{})

	for _, jobFunc := range jobs {
		wg.Add(1)
		out := make(chan interface{})

		go func(in, out chan interface{}) {
			defer wg.Done()
			defer close(out)

			jobFunc(in, out)
		}(in, out)

		in = out
	}

	wg.Wait()

}

// SingleHash
func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup

	for data := range in {
		wg.Add(1)

		strData := fmt.Sprintf("%v", data)
		md5Data := DataSignerMd5(strData)

		go func(data string, md5Data string, out chan interface{}) {
			defer wg.Done()

			crc32Chanel := make(chan string)
			crc32Md5Chanel := make(chan string)

			go calcCRC32(crc32Chanel, data)
			go calcCRC32(crc32Md5Chanel, md5Data)

			data = <-crc32Chanel
			md5Data = <-crc32Md5Chanel

			out <- data + "~" + md5Data

		}(strData, md5Data, out)
	}

	wg.Wait()
}

func calcCRC32(chanel chan string, data string) {
	result := DataSignerCrc32(data)
	chanel <- result
}

// MultiHash
func MultiHash(in, out chan interface{}) {
	var wgMultiHash sync.WaitGroup

	for data := range in {
		wgMultiHash.Add(1)
		singleHashhashData := fmt.Sprintf("%v", data)

		go func(data string, wgMultiHash *sync.WaitGroup) {
			defer wgMultiHash.Done()

			arrayResults := make([]string, 6)
			var wg6 sync.WaitGroup
			wg6.Add(6)

			for th := 0; th < 6; th++ {
				go func(th int, data string) {
					defer wg6.Done()
					arrayResults[th] = DataSignerCrc32(fmt.Sprintf("%d%s", th, data))
				}(th, data)
			}

			wg6.Wait()
			finalResult := strings.Join(arrayResults, "")
			out <- finalResult
		}(singleHashhashData, &wgMultiHash)

	}
	wgMultiHash.Wait()
}

// CombineResults
func CombineResults(in, out chan interface{}) {
	var results []string
	for data := range in {
		results = append(results, fmt.Sprintf("%v", data))
	}

	sort.Strings(results)
	finalResult := strings.Join(results, "_")
	out <- finalResult

}
