package main

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct {
	// mu    sync.Mutex
	// count int
	count int64
}

func (c *Counter) Increment() {
	// c.mu.Lock()
	// c.count = c.count + 25
	// fmt.Println(c.count)
	// c.mu.Unlock()

	atomic.AddInt64(&c.count, 5000000)

}

func (c *Counter) Value() int64 {
	// c.mu.Lock()
	// defer c.mu.Unlock()
	// return c.count
	return atomic.LoadInt64(&c.count)
}

// func before() {
// 	start := time.Now()
// 	c := Counter{}
// 	var wg sync.WaitGroup

// 	for i := 1; i < 1000000000; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			c.Increment()
// 		}()
// 	}
// 	wg.Wait()

// 	fmt.Println("Final Value: ", c.Value())

// 	finished := time.Since(start)
// 	fmt.Println("Program took: ", finished)

// }
func after() {
	start := time.Now()
	count := Counter{}
	numOfGoRoutines := runtime.NumCPU()
	fmt.Println("num cpus:", numOfGoRoutines)
	jobsPerGoRoutine := 1000000000 / numOfGoRoutines

	var wg sync.WaitGroup
	for i := 0; i < numOfGoRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < jobsPerGoRoutine; j++ {
				count.Increment()
			}
		}()
	}
	wg.Wait()

	fmt.Println("Final Value: ", count.Value())

	finished := time.Since(start)
	fmt.Println("Program took: ", finished)

}

func factorial(n int64) *big.Int {
	result := big.NewInt(1)
	for i := int64(2); i < n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

func main() {

	// after()

	n := int64(100)
	fact := factorial(n)
	factString := fact.String()
	fmt.Printf("The factorial of %d has %d digits \n", n, len(factString))
	fmt.Printf("First 100 digits of %d! is: \n%s\n", n, factString)

}
