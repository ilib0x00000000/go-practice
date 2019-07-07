package main

/**
 _   _          _   _         _______ _       _ _______  _______  _______  _______  _______  _______  _______  _______
(_) | |        (_) | |       / _____ \\\     /// _____ \/ _____ \/ _____ \/ _____ \/ _____ \/ _____ \/ _____ \/ _____ \
| | | |        | | | |       |/     \| \\   // |/     \||/     \||/     \||/     \||/     \||/     \||/     \||/     \|
| | | |        | | | |______ ||     ||  \\ //  ||     ||||     ||||     ||||     ||||     ||||     ||||     ||||     ||
| | | |        | | | |_____ \||     ||   ||    ||     ||||     ||||     ||||     ||||     ||||     ||||     ||||     ||
| | | |        | | | |     \|||     ||  //\\   ||     ||||     ||||     ||||     ||||     ||||     ||||     ||||     ||
| | | |_______ | | | |____ /||\_____/| //  \\  |\_____/||\_____/||\_____/||\_____/||\_____/||\_____/||\_____/||\_____/|
|_| |_________||_| |_|______/\_______///    \\ \_______/\_______/\_______/\_______/\_______/\_______/\_______/\_______/



解决
ld: malformed file
/Library/Developer/CommandLineTools/SDKs/MacOSX10.14.sdk/usr/lib/libSystem.tbd:4:18: error: unknown enumerated scalar

https://github.com/ilib0x00000000/go-practice/blob/master/kafka/demo.go

升级 xcode即可
*/

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go consumer()
	go asyncProducer()

	wg.Wait()
}

func consumer() {
	groupID := "group-1"

	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // 初始从最新的 offset 开始

	consumer, err := cluster.NewConsumer(strings.Split("localhost:9092", ","), groupID, []string{""}, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	go func() {
		errors := consumer.Errors()
		noti := consumer.Notifications()

		for {
			select {
			case err := <-errors:
				log.Println(err)
			case <-noti:
			}
		}
	}()

	for msg := range consumer.Messages() {
		fmt.Fprintf(os.Stdout, "topic: %s\n", msg.Topic)
		fmt.Fprintf(os.Stdout, "partition: %d\n", msg.Partition)
		fmt.Fprintf(os.Stdout, "offset: %d\n", msg.Offset)
		fmt.Fprintf(os.Stdout, "value: %s\n", msg.Value)

		// MarkOffset 不是实时写入 kafka
		// 有可能在程序 crash 时丢掉未提交的 offset
		consumer.MarkOffset(msg, "")
	}
}

// syncProducer 同步生产者
//并发量小的时候，可以使用这种方式
func syncProducer() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewCustomHashPartitioner(fnv.New32a)
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Timeout = 5 * time.Second
	config.Version = sarama.V2_1_0_0

	p, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	// 生产消息并发送
	http.HandleFunc("/create/comment", func(w http.ResponseWriter, r *http.Request) {
		access := fmt.Sprintf("time: %s, ip: %s, url: %s", time.Now().Format("2006-01-02 15:04:05"),
			r.RemoteAddr, r.URL.Path)

		msg := &sarama.ProducerMessage{
			Topic: "access-log",
			Value: sarama.ByteEncoder(access),
		}
		if _, _, err := p.SendMessage(msg); err != nil {
			log.Println(err)
			w.Write([]byte("System error"))
			return
		}

		w.Write([]byte("ok"))
	})

	http.ListenAndServe(":8080", nil)
}

// asyncProducer 异步生产者
// 并发量大的时候，必须使用的方式
func asyncProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // must
	config.Producer.Timeout = 5 * time.Second

	p, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	go func() {
		errors := p.Errors()
		success := p.Successes()

		for {
			select {
			case err := <-errors:
				if err != nil {
					log.Println(err)
				}
			case <-success:
			}
		}
	}()

	for {
		v := "async: " + strconv.Itoa(rand.New(rand.NewSource(time.Now().Unix())).Intn(10000))
		fmt.Fprintln(os.Stdout, v)

		msg := &sarama.ProducerMessage{
			Topic: "async-topic",
			Value: sarama.ByteEncoder(v),
		}

		p.Input() <- msg
		time.Sleep(1 * time.Second)
	}
}
