package rediscache

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Mutex 分布式锁
type RedisMutex struct {
	redisService *RedisService
	//命名一个名字
	lockkey string
	//锁的值
	lockvalue string
	//最多可以获取锁的时间，超过自动解锁
	expiry time.Duration
	//失败最多获取锁的次数
	tries int
	//获取锁失败后等待多少时间后重试
	delay time.Duration
	//当前尝试了多少次
	curTries int
}

// Lock ...
func (m *RedisMutex) RedisLock() bool {
	err := m.redisService.PutNxExStr(m.lockkey, m.lockvalue, int(m.expiry.Seconds()))
	if nil != err {
		if netError, ok := err.(net.Error); ok {
			slog.Error(fmt.Sprintf("[%v-%v]获取Redis锁失败,%s", m.lockkey, m.lockvalue, netError.Error()))
			return false
		}
		if m.curTries > m.tries {
			slog.Error(fmt.Sprintf("[%v-%v]获取Redis锁失败,当前第%d次获取,总次数%d", m.lockkey, m.lockvalue, m.curTries, m.tries))
			return false
		}
		slog.Debug(fmt.Sprintf("[%v-%v]第%d次获取锁失败,等待%dms后重试", m.lockkey, m.lockvalue, m.curTries, time.Duration(m.delay).Milliseconds()))
		m.curTries++
		time.Sleep(m.delay)
		return m.RedisLock()
	}
	return true
}

// ReleseLock ...
func (m *RedisMutex) RedisReleseLock() bool {
	conn := m.redisService.RedisPool.Get()
	defer conn.Close()
	_, err := redis.Int(RdisScriptDelKv.Do(conn, m.lockkey, m.lockvalue))
	return err == nil
}

// NewMutex ...
// 默认8秒过期，重试次数16次，失败500毫秒获取一次
func RedisNewMutex(lockkey, lockvalue string, redisService *RedisService, options ...RedisMutexOption) *RedisMutex {
	mutex := &RedisMutex{
		lockkey:      lockkey,
		lockvalue:    lockvalue,
		expiry:       8 * time.Second,
		tries:        16,
		delay:        500 * time.Millisecond,
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
func RedisMutexSetDelay(delay time.Duration) RedisMutexOption {
	return OptionFunc(func(m *RedisMutex) {
		m.delay = delay
	})
}
