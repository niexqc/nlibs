package nmq

import (
	"context"
	"log/slog"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

type NMqProduer struct {
	GroupName string
	Topic     string
	Producer  rocketmq.Producer
	Consumer  rocketmq.PushConsumer
}

type NMqProperty struct {
	Key string
	Val string
}

func NewNMqProduer(nameSvrAddr, topic, groupName string) *NMqProduer {
	rlog.SetLogLevel("warn")
	myProducer, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{nameSvrAddr}), // NameServer地址
		producer.WithRetry(0),                          // 重试次数
		producer.WithGroupName(groupName),
	)
	if err != nil {
		panic(err)
	}
	if err := myProducer.Start(); err != nil {
		panic("生产者启动失败: " + err.Error())
	}
	slog.Info("生产者已启动")
	return &NMqProduer{
		GroupName: groupName,
		Topic:     topic,
		Producer:  myProducer,
	}
}

// 发送顺序消息，tag可以为空
func (mq *NMqProduer) SendOrderMsg(msgContent, shardingKey, tag string, prp ...NMqProperty) (msgId string, err error) {
	msg := &primitive.Message{
		Topic: mq.Topic,
		Body:  []byte(msgContent),
	}
	msg.WithShardingKey(shardingKey)
	msg.WithTag(tag)
	if len(prp) > 0 {
		for _, v := range prp {
			msg.WithProperty(v.Key, v.Val)
		}
	}
	res, err := mq.Producer.SendSync(context.Background(), msg)
	if nil != err {
		return "", err
	}
	return res.MsgID, nil
}
