package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"coolpy7_benchmark/src/packet"
	"coolpy7_benchmark/src/transport"
	"github.com/juju/ratelimit"
)

// 一对一发布订阅测试工具
// 本工具建立一个订单及通过一个发布测试单一对一对信时最大通信质量

var urlString = flag.String("url", "tcp://127.0.0.1:1883", "broker url")
var workers = flag.Int("workers", 1, "number of workers")
var duration = flag.Int("duration", 30, "duration in seconds")
var publishRate = flag.Int("publish-rate", 0, "messages per second")
var receiveRate = flag.Int("receive-rate", 0, "messages per second")

var sent int32
var received int32
var delta int32
var total int32

var wg sync.WaitGroup

func main() {
	flag.Parse()

	fmt.Printf("Start benchmark of %s using %d workers for %d seconds.\n", *urlString, *workers, *duration)

	go func() {
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, syscall.SIGINT, syscall.SIGTERM)

		<-finish
		fmt.Println("Closing...")
		os.Exit(0)
	}()

	if int(*duration) > 0 {
		time.AfterFunc(time.Duration(*duration)*time.Second, func() {
			fmt.Println("Finishing...")
			os.Exit(0)
		})
	}

	wg.Add(*workers * 2)

	for i := 0; i < *workers; i++ {
		id := strconv.Itoa(i)

		go consumer(id)
		go publisher(id)
	}

	go reporter()

	wg.Wait()
}

func connection(id string) transport.Conn {
	conn, err := transport.Dial(*urlString)
	if err != nil {
		panic(err)
	}

	mqttURL, err := url.Parse(*urlString)
	if err != nil {
		panic(err)
	}

	connect := packet.NewConnectPacket()
	connect.ClientID = "benchmark/" + id

	if mqttURL.User != nil {
		connect.Username = mqttURL.User.Username()
		pw, _ := mqttURL.User.Password()
		connect.Password = pw
	}

	err = conn.Send(connect)
	if err != nil {
		panic(err)
	}

	pkt, err := conn.Receive()
	if err != nil {
		panic(err)
	}

	if connack, ok := pkt.(*packet.ConnackPacket); ok {
		if connack.ReturnCode == packet.ConnectionAccepted {
			fmt.Printf("Connected: %s\n", id)

			return conn
		}
	}

	panic("connection failed")
}

func consumer(id string) {
	name := "consumer/" + id
	conn := connection(name)

	subscribe := packet.NewSubscribePacket()
	subscribe.ID = 1
	subscribe.Subscriptions = []packet.Subscription{
		{Topic: id, QOS: 0},
	}

	err := conn.Send(subscribe)
	if err != nil {
		panic(err)
	}

	var bucket *ratelimit.Bucket
	if *receiveRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*receiveRate), int64(*receiveRate))
	}

	for {
		if bucket != nil {
			bucket.Wait(1)
		}

		_, err := conn.Receive()
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&received, 1)
		atomic.AddInt32(&delta, -1)
		atomic.AddInt32(&total, 1)
	}
}

func publisher(id string) {
	name := "publisher/" + id
	conn := connection(name)

	publish := packet.NewPublishPacket()
	publish.Message.Topic = id
	publish.Message.Payload = []byte("foofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofofoofoofoofoofoofoofo")

	var bucket *ratelimit.Bucket
	if *publishRate > 0 {
		bucket = ratelimit.NewBucketWithRate(float64(*publishRate), int64(*publishRate))
	}

	for {
		if bucket != nil {
			bucket.Wait(1)
		}

		err := conn.BufferedSend(publish)
		if err != nil {
			panic(err)
		}

		atomic.AddInt32(&sent, 1)
		atomic.AddInt32(&delta, 1)
	}
}

func reporter() {
	var iterations int32

	for {
		time.Sleep(1 * time.Second)

		curSent := atomic.LoadInt32(&sent)
		curReceived := atomic.LoadInt32(&received)
		curDelta := atomic.LoadInt32(&delta)
		curTotal := atomic.LoadInt32(&total)

		iterations++

		fmt.Printf("Sent: %d msgs - ", curSent)
		fmt.Printf("Received: %d msgs ", curReceived)
		fmt.Printf("(Buffered: %d msgs) ", curDelta)
		fmt.Printf("(Average Throughput: %d msg/s)\n", curTotal/iterations)

		atomic.StoreInt32(&sent, 0)
		atomic.StoreInt32(&received, 0)
	}
}
