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

func testPing2() {
	tm := time.Now()

	count := 0
	ch := make(chan ping.Result)
	for i := 100; i < 101; i++ {
		for j := 100; j < 200; j++ {
			for k := 100; k < 200; k++ {
				count++
				ip := fmt.Sprintf("60.%d.%d.%d", i, j, k)
				param := ping.Param{Addr: ip, DataLen: 32, Segment: true, Timeout: 1}
				go func() {
					r, _ := ping.Ping(param)
					/*if e != nil {
						fmt.Printf("ping %s ret error %s\n", param.Addr, e.Error())
					} else {
						fmt.Printf("ping %s cost %d ms, ttl %d\n", param.Addr, r.Time, r.Ttl)
					}*/

					ch <- r
				}()
			}
		}
	}

	for i := 1; i < count; i++ {
		<-ch
		//fmt.Printf("ping ret %v\n", r)
	}

	cost := time.Since(tm) / 1000 / 1000

	fmt.Printf("test completed, ping %d ip cost %d ms\n", count, cost)
}

func main() {
	//testPing()
	testPing2()

	println("exit main")
}
