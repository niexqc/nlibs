package nrocketmq

import (
	"context"
	"log/slog"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

type NRocketMq struct {
	GroupName   string
	Topic       string
	Tag         string
	ShardingKey string // 顺序消息的key
	Producer    rocketmq.Producer
	Consumer    rocketmq.PushConsumer
}

func NewNRocketMq(nameSvrAddr, groupName, topic, tag, shardingKey string) *NRocketMq {
	rlog.SetLogLevel("error")
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

	myCunsumer, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{nameSvrAddr}),
		consumer.WithGroupName(groupName),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithMaxReconsumeTimes(16),         // 死信队列，重试16次
		consumer.WithConsumeMessageBatchMaxSize(1), //单条消费
	)

	return &NRocketMq{
		GroupName:   groupName,
		Topic:       topic,
		Tag:         tag,
		ShardingKey: shardingKey,
		Producer:    myProducer,
		Consumer:    myCunsumer,
	}
}

// 订阅消息
func (mq *NRocketMq) Subscribe(onMsg func(ctx context.Context, msg *primitive.MessageExt) (consumer.ConsumeResult, error)) {
	err := mq.Consumer.Subscribe(mq.Topic, consumer.MessageSelector{Type: consumer.TAG, Expression: mq.Tag}, func(ctx context.Context, imsgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		return onMsg(ctx, imsgs[0])
	})
	if nil != err {
		panic(err)
	}
	if err := mq.Consumer.Start(); err != nil {
		panic("消费者启动失败: " + err.Error())
	}
	slog.Info("消费者已启动")
}

// 发送顺序消息
func (mq *NRocketMq) SendOrderMsg(msgContent string) (msgId string, err error) {
	msg := &primitive.Message{
		Topic: mq.Topic,
		Body:  []byte(msgContent),
	}
	msg.WithShardingKey(mq.ShardingKey)
	msg.WithTag(mq.Tag)
	res, err := mq.Producer.SendSync(context.Background(), msg)
	if nil != err {
		return "", err
	}
	return res.MsgID, nil
}
