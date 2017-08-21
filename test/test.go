package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jnscode/nettool/ping"
)

func testPing() {
	param := ping.Param{"www.baidu.com", 32, true, 5}
	r, e := ping.Ping(param)
	if e != nil {
		fmt.Printf("ping %s ret error\n", param.Addr, e.Error())
	} else {
		fmt.Printf("ping %s ret %v\n", param.Addr, r)
	}
}

func testPing2() {
	tm := time.Now()

	ch := make(chan ping.Result)
	for i := 1; i < 255; i++ {
		ip := "192.168.1." + strconv.Itoa(i)
		param := ping.Param{ip, 32, true, 5}
		go func() {
			r, e := ping.Ping(param)
			if e != nil {
				fmt.Printf("ping %s ret error\n", param.Addr, e.Error())
			} else {
				fmt.Printf("ping %s ret %v\n", param.Addr, r)
			}

			ch <-r
		}()
	}

	for i := 1; i < 255; i++ {
		r := <-ch
		fmt.Printf("ping ret %v\n", r)
	}

	cost := time.Since(tm) / 1000 / 1000

	fmt.Printf("test completed, cost %d ms", cost)
}

func main() {
	testPing()
	testPing2()

	println("exit main")
}
