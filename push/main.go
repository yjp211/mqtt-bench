package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Dict map[string]interface{}
type List []interface{}

var (
	C          = flag.Int("c", 2, "Number of multipe requests to make")
	N          = flag.Int("n", 10, "Number of clients to test")
	D          = flag.Bool("d", false, "message diffuse")
	Host       = flag.String("h", "127.0.0.1", "host addr")
	Port       = flag.Int("p", 9999, "port")
	UserPerfix = flag.String("u", "GoClient", "user perfix")
	MesgPerfix = flag.String("m", "Hello", "message perfix")
	Topic      = flag.String("t", "test", "topic key")
)

func pushMesg(i int, mesg string) {
	params := Dict{
		"diffuse": *D,
		"msg":     mesg,
		"topic":   *Topic,
		"weight":  2,
		"version": 1,
	}
	url := fmt.Sprintf("http://%s:%d/provider/v1/publish", *Host, *Port)
	HttpPostJson(url, params, "mqtt-bench", "123@.root")
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := *C
	n := *N
	g := n / c
	var wg sync.WaitGroup
	wg.Add(c)

	t1 := time.Now()
	for i := 0; i < c; i++ {
		go func(i int) {
			for j := 0; j < g; j++ {
				id := i*g + j
				content := fmt.Sprintf("%s_%d", *MesgPerfix, id)
				pushMesg(id, content)
				time.Sleep(time.Second)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	t2 := time.Now()
	useTime := t2.Sub(t1).Seconds()
	qps := float64(n) / useTime
	fmt.Printf("use time:%f, qps:%f \n", useTime, qps)

}
