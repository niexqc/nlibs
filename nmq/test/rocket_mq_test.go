package nrocketmq_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/niexqc/nlibs/nmq"
	"github.com/niexqc/nlibs/ntools"
)

var NRocketMq *nmq.NMqConsumer
var Topic = "t001"
var NameServer = "192.168.0.253:9876"

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}

func onReciveMsg(ctx context.Context, msg *primitive.MessageExt) (consumer.ConsumeResult, error) {

	slog.Info(msg.GetProperty("type"))
	slog.Info(msg.GetTags())

	slog.Info(string(msg.Body))
	return consumer.ConsumeSuccess, nil
}

func TestSendMsg(t *testing.T) {
	p01 := nmq.NewNMqProduer(NameServer, Topic, "p01")
	msgId, _ := p01.SendOrderMsg("p01-test01"+ntools.Time2Str(time.Now()), "o", "tA", nmq.NMqProperty{Key: "type", Val: "AAA"})
	fmt.Println(msgId)
	// p02 := nmq.NewNMqProduer(NameServer, Topic, "p02")
	// p02.SendOrderMsg("p02-test01"+ntools.Time2Str(time.Now()), "o", "")

	c01 := nmq.NewNMqConsumer(NameServer, Topic, "cc02", false)
	c01.Subscribe("*", onReciveMsg)

	time.Sleep(2 * time.Second)
}
