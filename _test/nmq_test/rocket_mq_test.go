package nrocketmq_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/niexqc/nlibs/nmq"
	"github.com/niexqc/nlibs/ntools"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}

func TestNmqProducerAndConsumer(t *testing.T) {
	//如果测试未通过，先检查NameServer是否正确
	var Topic = "t001"
	var NameServer = "192.168.0.253:9876"

	producer := nmq.NewNMqProduer(NameServer, Topic, "p01")
	msgId, err := producer.SendOrderMsg("p01-test01"+ntools.Time2Str(time.Now()), "o", "tA", nmq.NMqProperty{Key: "type", Val: "AAA"})
	ntools.TestErrPainic(t, "NMqProduer发送测试", err)
	slog.Info("NMqProduer SendOrderMsg ", "msgId", msgId)

	recv := make(chan string, 1)

	onReciveMsg := func(ctx context.Context, msg *primitive.MessageExt) (consumer.ConsumeResult, error) {
		slog.Info(string(msg.Body) + " tag:" + msg.GetTags() + " Type:" + msg.GetProperty("type"))
		recv <- msg.MsgId
		return consumer.ConsumeSuccess, nil
	}
	c01 := nmq.NewNMqConsumer(NameServer, Topic, "cc02", false)
	c01.Subscribe("*", onReciveMsg)
	recvMsgId := <-recv
	ntools.TestStrContains(t, "NMqConsumer接收刚发送的消息", msgId, recvMsgId)
}
