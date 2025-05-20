package ntools_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ntools"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)

}

func TestHttpClientPool(t *testing.T) {
	client := ntools.NewNHttpClientPool(10, 10, 30*time.Second, 30*time.Second)
	req, _ := http.NewRequest("GET", "http://www.baidu.com", nil)

	client.RunRequst(req, func(resp *http.Response, err error) {
		data, _ := io.ReadAll(resp.Body)
		slog.Debug(string(data))
	})

	time.Sleep(30 * time.Second)
}
