package nrocketmq_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/niexqc/nlibs/nrocketmq"
	"github.com/niexqc/nlibs/ntools"
)

var NRocketMq *nrocketmq.NRocketMq

func init() {

	ntools.SlogConf("test", "debug", 1, 2)
	NRocketMq = nrocketmq.NewNRocketMq("192.168.0.253:9876", "g001", "t001", "tag001", "s001")
}

func onReciveMsg(ctx context.Context, msg *primitive.MessageExt) (consumer.ConsumeResult, error) {
	slog.Info(string(msg.Body))
	return consumer.ConsumeSuccess, nil
}

func TestMq(t *testing.T) {
	go func() {
		for i := range 1000 {
			NRocketMq.SendOrderMsg(fmt.Sprintf("消息线程1编号:%d", i))
			time.Sleep(20 * time.Millisecond)
		}
	}()
	go func() {
		for i := range 1000 {
			NRocketMq.SendOrderMsg(fmt.Sprintf("消息线程2编号:%d", i))
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go func() {
		for i := range 1000 {
			NRocketMq.SendOrderMsg(fmt.Sprintf("消息线程3编号:%d", i))
			time.Sleep(300 * time.Millisecond)
		}
	}()
	NRocketMq.Subscribe(onReciveMsg)
	time.Sleep(10 * time.Second)
}
