package nmq

import (
	"context"
	"log/slog"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/niexqc/nlibs/ntools"
)

type NMqConsumer struct {
	GroupName string
	Topic     string
	Consumer  rocketmq.PushConsumer
}

func NewNMqConsumer(nameSvrAddr, topic, groupName string, broadCastingMode bool) *NMqConsumer {
	rlog.SetLogLevel("error")
	consumerModel := ntools.If3(broadCastingMode, consumer.BroadCasting, consumer.Clustering)
	myCunsumer, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{nameSvrAddr}),
		consumer.WithGroupName(groupName),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithMaxReconsumeTimes(16),         // 死信队列，重试16次
		consumer.WithConsumeMessageBatchMaxSize(1), //单条消费
		consumer.WithConsumerModel(consumerModel),
	)

	return &NMqConsumer{
		GroupName: groupName,
		Topic:     topic,
		Consumer:  myCunsumer,
	}
}

// 订阅消息,tag可以为空
func (mq *NMqConsumer) Subscribe(tag string, onMsg func(ctx context.Context, msg *primitive.MessageExt) (consumer.ConsumeResult, error)) {
	msgSelector := consumer.MessageSelector{}
	if tag != "" {
		msgSelector = consumer.MessageSelector{Type: consumer.TAG, Expression: tag}
	}
	err := mq.Consumer.Subscribe(mq.Topic, msgSelector, func(ctx context.Context, imsgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		return onMsg(ctx, imsgs[0])
	})
	if nil != err {
		panic(err)
	}
	if err := mq.Consumer.Start(); err != nil {
		panic("消费者启动失败: " + err.Error())
	}
	slog.Info("消费者订阅成功")
}
