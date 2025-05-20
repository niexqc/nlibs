package ntools

import (
	"net/http"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type NHttpClientPool struct {
	HttpClientPool *sync.Pool
	AntsPool       *ants.Pool
}

func NewNHttpClientPool(maxConcurrent, maxIdleConns int, httpTimeout, idleConnTimeout time.Duration) *NHttpClientPool {
	syncPool := &sync.Pool{
		New: func() interface{} {
			return &http.Client{
				Timeout: httpTimeout,
				Transport: &http.Transport{
					MaxIdleConns:    maxIdleConns,
					MaxConnsPerHost: maxConcurrent,
					IdleConnTimeout: idleConnTimeout,
				},
			}
		},
	}
	antsPool, _ := ants.NewPool(maxConcurrent, ants.WithNonblocking(false))
	o := &NHttpClientPool{HttpClientPool: syncPool, AntsPool: antsPool}
	return o
}

func (hcp *NHttpClientPool) RunRequst(request *http.Request, callback func(resp *http.Response, err error)) {
	hcp.AntsPool.Submit(func() {
		httpClient := hcp.HttpClientPool.Get().(*http.Client)
		defer hcp.HttpClientPool.Put(httpClient)

		resp, err := httpClient.Do(request)
		defer func() {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}()
		callback(resp, err)
	})
}
