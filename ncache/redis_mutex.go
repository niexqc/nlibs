package ncache

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/niexqc/nlibs/ntools"
)

// Mutex 分布式锁
type RedisMutex struct {
	redisService *RedisService
	//命名一个名字
	name string
	//最多可以获取锁的时间，超过自动解锁
	expiry time.Duration
	//失败最多获取锁的次数
	tries int
	//获取锁失败后等待多少时间后重试
	delay time.Duration
	//锁的值
	value string
	//当前尝试了多少次
	curTries int
}

// Lock ...
func (m RedisMutex) Lock() bool {
	err := m.redisService.PutNxExStr(m.name, m.value, int(m.expiry.Seconds()))
	if nil != err {
		if netError := err.(net.Error); netError != nil {
			slog.Error(fmt.Sprintf("获取Redis锁失败,%s", netError.Error()))
			return false
		}
		if m.curTries > m.tries {
			slog.Error(fmt.Sprintf("获取Redis锁失败,当前第%d次获取,总次数%d\n", m.curTries, m.tries))
			return false
		}
		m.curTries++
		time.Sleep(m.delay)
		return m.Lock()
	}
	return true
}

// ReleseLock ...
func (m RedisMutex) ReleseLock() bool {
	conn := m.redisService.RedisPool.Get()
	defer conn.Close()
	_, err := redis.Int(RdisScriptDelKv.Do(conn, m.name, m.value))
	return err == nil
}

// NewMutex ...
func RedisNewMutex(name string, redisService *RedisService, options ...RedisMutexOption) *RedisMutex {
	mutex := &RedisMutex{
		name:         name,
		expiry:       8 * time.Second,
		tries:        16,
		delay:        500 * time.Millisecond,
		value:        ntools.UUIDStr(true),
		curTries:     1,
		redisService: redisService,
	}
	for _, option := range options {
		option.Apply(mutex)
	}
	return mutex
}

// An Option configures a mutex.
type RedisMutexOption interface {
	Apply(*RedisMutex)
}

// Apply 实现Option.Apply
func (f OptionFunc) Apply(mutex *RedisMutex) {
	f(mutex)
}

// OptionFunc 配置方法
type OptionFunc func(*RedisMutex)

// SetExpiry 设置 锁最多可以占用的时间，超过自动解锁
func RedisMutexSetExpiry(expiry time.Duration) RedisMutexOption {
	return OptionFunc(func(m *RedisMutex) {
		m.expiry = expiry
	})
}

// SetTries 设置失败最多获取锁的次数
func RedisMutexSetTries(tries int) RedisMutexOption {
	return OptionFunc(func(m *RedisMutex) {
		m.tries = tries
	})
}

// SetDelay 设置获取锁失败后等待多少时间后重试
func RedisMutexSetDelay(expiry time.Duration) RedisMutexOption {
	return OptionFunc(func(m *RedisMutex) {
		m.expiry = expiry
	})
}

// SetValue 设置锁的值
func RedisMutexSetValue(tries int) RedisMutexOption {
	return OptionFunc(func(m *RedisMutex) {
		m.tries = tries
	})
}
