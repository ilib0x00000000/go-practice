package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kafka demo producer"
	app.Usage = "go run main.go"
	app.Version = "0.0.1"
	app.Flags = args()
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Action = action
	app.Run(os.Args)
}


// args 命令行参数定义
func args() []cli.Flag {
	return []cli.Flag {
		cli.StringFlag{
			Name:   "listen-address, la",
			Value:  "127.0.0.1:9000",
			Usage:  "Listen address for api",
			EnvVar: "LISTEN_ADDR",
		},
		cli.StringFlag{
			Name:   "kafka-brokers, kb",
			Value:  "kafka1:19092,kafka2:29092,kafka3:39092",
			Usage:  "kafka brokers in comma separated value",
			EnvVar: "KAFKA_BROKERS",
		},
		cli.StringFlag{
			Name:   "kafka-topic, kt",
			Value:  "hello",
			Usage:  "kafka topic to push",
			EnvVar: "KAFKA_TOPIC",
		},
	}
}


// action 创建 kafka 生产者 并启动路由服务
func action(c *cli.Context) error {
	log.Println("kafka demo: producer")
	log.Println("(c) ilib0x00000000 2019")

	listenAddr := c.String("listen-address")
	kafkaBrokers := c.String("kafka-brokers")
	topic := c.String("kafka-topic")

	log.Println("listen address: ", listenAddr)
	log.Println("kafka brokers: ", kafkaBrokers)
	log.Println("kafka topic: ", topic)

	brokerUrls := strings.Split(kafkaBrokers, ",")

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewCustomHashPartitioner(fnv.New32a)
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_1_0_0

	log.Println("start make topic")
	err := createTopic(config, brokerUrls[0], topic)
	if err != nil {
		log.Printf("create topic error: %#v", err)
		return err
	}

	log.Println("start make producer")
	producer, err := sarama.NewAsyncProducer(brokerUrls, config)
	if err != nil {
		log.Printf("create producer error: %#v", err)
		return err
	}
	defer producer.Close()

	log.Println("start server ...")
	errChan := make(chan error, 1)
	go func() {
		errChan <- server(producer, topic, listenAddr)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

LOOP:
	for {
		select {
		case succ := <-producer.Successes():
			log.Println("**********************************************")
			log.Println("生产者发送消息成功")
			log.Println("key: ", succ.Key)
			log.Println("value: ", succ.Value)
			log.Println("offset: ", succ.Offset)
			log.Println("timestamp: ", succ.Timestamp)
			log.Println("partition: ", succ.Partition)
			log.Println("**********************************************")
		case fail := <- producer.Errors():
			log.Println("xxxxxxxxxxxxxxxxxxxxxxxxxx")
			log.Println("生产者发送消息失败")
			log.Println(fail.Error())
			log.Println(fail.Err.Error())
		case <-signals:
			log.Println("got an interrupt signal, exit")
			break LOOP
		case err := <- errChan:
			log.Printf("error: %#v", err)
			log.Println("errored when running api, exit ...")
			break LOOP
		}
	}

	return nil
}


// createTopic 创建 topic
func createTopic(config *sarama.Config, brokerURL string, topicName string) error {
	broker := sarama.NewBroker(brokerURL)
	broker.Open(config)
	status, err := broker.Connected()
	if err != nil {
		log.Printf("broker connect failed: %#v", err)
		return err
	}
	log.Println("broker connect status: ", status)

	details := map[string]*sarama.TopicDetail{
		topicName: &sarama.TopicDetail{
			NumPartitions:     2,
			ReplicationFactor: 1,
			ConfigEntries:     make(map[string]*string),
		},
	}

	request := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: details,
	}

	response, err := broker.CreateTopics(&request)
	if err != nil {
		log.Printf("create topics fail: %#v", err)
		return err
	}

	log.Printf("response length: %d", len(response.TopicErrors))
	for k, v := range response.TopicErrors {
		log.Printf("key is %s. value is %#v. val msg: %#v", k, v.Err.Error(), v.ErrMsg)
	}
	log.Printf("create topics response: %#v", response)

	broker.Close()
	return nil
}


// server 启动 HTTP 服务
func server(producer sarama.AsyncProducer, topicName string, listenAddr string) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/api/v1/data", func(ctx *gin.Context) {
		parent := context.Background()
		defer parent.Done()

		form := &struct {
			Text string `form:"text" json:"text"`
		}{}

		err := ctx.ShouldBindJSON(form)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": map[string]interface{}{
					"message": fmt.Sprintf("error while bind request param: %s", err.Error()),
				},
			})
			ctx.Abort()
			return
		}

		formInBytes, err := json.Marshal(form)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"error": map[string]interface{}{
					"message": fmt.Sprintf("error while marshalling json: %s", err.Error()),
				},
			})
			ctx.Abort()
			return
		}

		// send message to kafka
		msg := &sarama.ProducerMessage{
			Topic:     topicName,
			Key:       sarama.StringEncoder(form.Text),
			Value:     sarama.ByteEncoder(formInBytes),
			Timestamp: time.Now(),
		}
		producer.Input() <- msg

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "success push data into kafka",
			"data":    form,
		})
	})

	return router.Run(listenAddr)
}
