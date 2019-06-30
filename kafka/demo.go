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

kafka入门：
	不出意外，第一个demo挂了

go build 出现错误：
`
# github.com/DataDog/zstd
ld: malformed file
/Library/Developer/CommandLineTools/SDKs/MacOSX10.14.sdk/usr/lib/libSystem.tbd:4:18: error: unknown enumerated scalar
platform:        zippered
                 ^~~~~~~~
 file '/Library/Developer/CommandLineTools/SDKs/MacOSX10.14.sdk/usr/lib/libSystem.tbd'
clang: error: linker command failed with exit code 1 (use -v to see invocation)
`
环境 Mac mojave 10.14.5
golang go1.12.5 darwin/amd64
*/

import (
	"fmt"
	"log"

	"github.com/sarama"
)

func main() {
	// 新建一个实例
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll // 等待服务器所有副本都保存成功后的响应
	// config.Producer.Partitioner = sarama.NewRandomPartitioner() // 随机向partition发送消息
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_2_0_0

	// 新建一个同步生产者
	client, err := sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, config)
	if err != nil {
		log.Fatal("producer closed, err: ", err)
		return
	}
	defer client.Close()

	// 定义一个生产消息
	msg := &sarama.ProducerMessage{}
	msg.Topic = "demo"
	msg.Key = sarama.StringEncoder("hello")
	msg.Value = sarama.StringEncoder("world")

	// 发送消息
	pid, offset, err := client.SendMessage(msg)
	if err != nil {
		fmt.Println("send failed, err: ", err)
	} else {
		fmt.Println("pid: ", pid)
		fmt.Println("offset: ", offset)
	}

	msg = &sarama.ProducerMessage{
		Topic: "demo",
		Key:   sarama.StringEncoder("xxx"),
		Value: sarama.StringEncoder("yyy"),
	}
	pid, offset, err = client.SendMessage(msg)
	if err != nil {
		fmt.Println("send failed, err: ", err)
	} else {
		fmt.Println("pid: ", pid)
		fmt.Println("offset: ", offset)
	}

	fmt.Println("send message succ")
}
