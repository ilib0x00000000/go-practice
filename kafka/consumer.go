// 参考 https://github.com/ErikJiang/kafka_cluster_example

package customer

import (
	"encoding/json"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/prometheus/common/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kafka demo consumer"
	app.Usage = "go run main.go"
	app.Version = "0.0.1"
	app.Flags = args()

	sort.Sort(cli.FlagsByName(app.Flags))
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

	log.Debug("args: ", os.Args)
}


// args 命令行参数定义
func args() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "kafka-brokers, kb",
			Value:  "kafka1:19092,kafka2:29092,kafka3:39092",
			Usage:  "kafka brokers in comma separated value",
			EnvVar: "KAFKA_BROKERS",
		},
		cli.StringFlag{
			Name:   "kafka-consumer-group, kcg",
			Value:  "consumer-group",
			Usage:  "kafka consumer group",
			EnvVar: "KAFKA_CONSUMER_GROUP_ID",
		},
		cli.StringFlag{
			Name:   "kafka-topic, kt",
			Value:  "hello",
			Usage:  "kafka topic to push",
			EnvVar: "KAFKA_TOPIC",
		},
	}
}


// action 创建 kafka 生产者并启动路由服务
func action(c *cli.Context) error {
	log.Info("kafka tutorial consume")
	log.Info("(c) ilib0x00000000 2019")

	brokerUrls := c.String("kafka-brokers")
	topic := c.String("kafka-topic")
	consumerGroup := c.String("kafka-consumer-group")

	log.Info("kafka-brokers: ", brokerUrls)
	log.Info("kafka-topic: ", topic)
	log.Info("kafka-consumer-group: ", consumerGroup)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go clusterConsumer(wg, strings.Split(brokerUrls, ","), []string{topic}, consumerGroup)
	wg.Wait()

	return nil
}


// clusterConsumer 支持 brokers cluster 的消费者
func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupID string) {
	defer wg.Done()

	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Version = sarama.V2_1_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 初始化消费者
	consumer, err := cluster.NewConsumer(brokers, groupID, topics, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// 捕获终止中断信号触发程序退出
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// 消费错误信息
	go func() {
		for err := range consumer.Errors() {
			log.Debugf("%s: Error: %#v\n", groupID, err)
		}
	}()

	// 消费通知信息
	go func() {
		for noti := range consumer.Notifications() {
			log.Debugf("%s: rebalanced: %#v \n", groupID, noti)
		}
	}()

	// 消费信息及监听信号
	var successes int

Loop:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				value := struct {
					Text string `form:"text" json:"text"`
				}{}

				err := json.Unmarshal(msg.Value, &value)
				if err != nil {
					log.Errorf("consume message json format error: %#v", err)
					break Loop
				}

				// 收到的信息
				log.Debugf("GroupID: %s, Topic: %s, Partition: %d, Offset: %d, Key: %s, Value: %s",
					groupID, msg.Topic, msg.Partition, msg.Offset, msg.Key, value.Text)
				consumer.MarkOffset(msg, "") // 标记信息为 已处理
				successes++
			}
		case <-signals:
			break Loop
		}
	}

	log.Debugf("%s consume %d messages", groupID, successes)
}
