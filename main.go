package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	C          = flag.Int("c", 2, "Number of multipe requests to make")
	N          = flag.Int("n", 10, "Number of clients to test")
	Host       = flag.String("h", "127.0.0.1", "host addr")
	Port       = flag.Int("p", 1883, "port")
	UserPerfix = flag.String("up", "GoClient", "user perfix")
	Topic      = flag.String("t", "test", "topic string")
	Account    = flag.String("a", "Guest", "user perfix")
	Passwd     = flag.String("pw", "Guest", "user perfix")
)

type Dict map[string]interface{}
type List []interface{}

var count int64 = 0

func parsePublish(user, clientId string, packet *Packet) int {
	packet.bodyPos = 0

	topic := packet.readString()
	msgId := atomic.AddInt64(&count, 1)
	message := string(packet.readBytes(int(packet.remainLength -
		packet.bodyPos)))

	fmt.Printf("[%d]:: user<%s> receive publish: topic<%v>, mesg<%v> for <%v>\n", msgId, user,
		topic, message, clientId)

	return 0

}

func connect(user string, addr string, port int) error {
	client := NewBrokerConn(addr, port, 0)
	if nil != client {
		clientId := user
		//	clientId = "0727eb5851ea415fac15c7f776d95e7b"
		account := *Account
		passwd := *Passwd
		topic := *Topic
		client.Auth(clientId, account, passwd)
		client.Subscribe(topic)
		for {
			packet, err := ReceivePacket(client.conn)
			if err == nil && packet != nil {
				if packet.command == PUBLISH_COMMAND {
					parsePublish(user, clientId, packet)
					//	client.PublishAck(msgId)
				}
			} else {
				fmt.Printf("-->user:<%s>, receive packet failed, %v\n", user, err)
				break
			}
		}
	} else {
		fmt.Printf("-->%s connect failed\n", user)
	}
	return nil
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := *C
	n := *N
	g := n / c

	t1 := time.Now()
	port := *Port
	addr := *Host
	for i := 0; i < c; i++ {
		for j := 0; j < g; j++ {
			id := i*g + j
			user := fmt.Sprintf("%s_%d", *UserPerfix, id)
			if id%2000 == 0 {
				time.Sleep(time.Second)
			}
			go connect(user, addr, port)
		}
	}
	t2 := time.Now()
	useTime := t2.Sub(t1).Seconds()
	qps := float64(n) / useTime
	fmt.Printf("use time:%f, qps:%f \n", useTime, qps)

	time.Sleep(2 * time.Second)

	var input string
	for {
		fmt.Scanln(&input)
	}

}
