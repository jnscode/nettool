package main

import (
	"fmt"
	"time"

	"github.com/jnscode/nettool/ping"
)

func testPing() {
	param := ping.Param{Addr: "www.baidu.com", DataLen: 32, Segment: true, Timeout: 5}
	r, e := ping.Ping(param)
	if e != nil {
		fmt.Printf("ping %s ret error %s\n", param.Addr, e.Error())
	} else {
		fmt.Printf("ping %s cost %d ms, ttl %d\n", param.Addr, r.Time, r.Ttl)
	}
}

func testPingPerform() {
	tm := time.Now()

	count := 0
	errCnt := 0
	ch := make(chan ping.Result)
	for i := 1; i <= 10; i++ {
		for j := 1; j <= 100; j++ {
			for k := 1; k <= 100; k++ {
				ip := fmt.Sprintf("60.%d.%d.%d", i, j, k)
				param := ping.Param{Addr: ip, DataLen: 32, Segment: true, Timeout: 1}
				go func() {
					r, _ := ping.Ping(param)
					ch <- r
				}()

				count++
			}
		}
	}

	for i := 1; i < count; i++ {
		r := <-ch
		if !r.Succ {
			errCnt++
		}
	}

	cost := time.Since(tm) / 1000 / 1000

	fmt.Printf("test completed, ping %d ip cost %d ms, error count %d\n", count, cost, errCnt)
}

func main() {
	testPing()
	testPingPerform()

	println("exit main")
}
