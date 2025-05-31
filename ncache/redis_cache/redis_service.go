package rediscache

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/niexqc/nlibs/nerror"
)

// RedisService ...
type RedisService struct {
	RedisPool *redis.Pool
}

// Int64自增
func (service *RedisService) Int64Incr(key string, expireMillisecond int64) (num int64, err error) {
	conn := service.RedisPool.Get()
	defer conn.Close()
	resp, err := redis.Int64(RdisScriptIntIncr.Do(conn, key, expireMillisecond))
	return resp, err
}

// GetStr ...
func (service *RedisService) GetStr(key string) (string, error) {
	conn := service.RedisPool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	return val, nil
}

// PutStr ...
func (service *RedisService) PutStr(key string, val string) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("SET", key, val))
	if nil != err {
		return err
	}
	if resp != "OK" {
		return errors.New("未返回OK")
	}
	return nil
}

// EXISTS ...
func (service *RedisService) Exist(key string) (bool, error) {
	conn := service.RedisPool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// ExistNoErr ...
func (service *RedisService) ExistWithoutErr(key string) bool {
	vexist, _ := service.Exist(key)
	return vexist
}

// 设置键值对并指定过期时间（​​原子性操作​​）,无论键是否存在，都会​​覆盖旧值​​并设置新的过期时间
func (service *RedisService) PutExStr(key string, val string, sencond int) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("SETEX", key, sencond, val))
	if nil != err {
		return err
	}
	if resp != "OK" {
		return errors.New("未返回OK")
	}
	return nil
}

// 仅在键​​不存在​​时设置键值对（​​原子性操作​​）
func (service *RedisService) PutNxExStr(key string, val string, sencond int) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("SET", key, val, "EX", sencond, "NX"))
	if nil != err {
		return err
	}
	if resp != "OK" {
		return errors.New("未返回OK")
	}
	return nil
}

// KeySetExpire ...
func (service *RedisService) KeySetExpire(key string, sencond int) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	resp, err := redis.Int64(conn.Do("EXPIRE", key, sencond))
	if err != nil {
		return err
	}
	if resp <= 0 {

		err := nerror.NewRunTimeError("设置成功数小于0")
		return err
	}
	return nil
}

// ClearKey 清理KEY
func (service *RedisService) ClearKey(key string) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	_, err := redis.Int(conn.Do("DEL", key))
	return err
}

// ClearByKeyPrefix 清理指定前缀的KEY
func (service *RedisService) ClearByKeyPrefix(keyPrefix string) (int, error) {
	conn := service.RedisPool.Get()
	defer conn.Close()
	keyPattner := fmt.Sprintf("%s*", keyPrefix)
	//扫描Key
	keys, err := scanKeysWithConn(conn, 0, keyPattner, nil, 1000)
	if nil != err {
		return 0, err
	}
	//删除
	if len(keys) > 0 {
		var delKeys = make([]interface{}, len(keys))
		for key := range keys {
			delKeys[key] = keys[key]
		}
		return redis.Int(conn.Do("DEL", delKeys...))
	}
	return 0, nil
}

// ...
func scanKeysWithConn(conn redis.Conn, cur int, keyPattner string, lastKeys []string, maxLen int) ([]string, error) {
	reply, err := conn.Do("SCAN", cur, "MATCH", keyPattner, "COUNT", maxLen)
	if nil == err {
		replyArray := reply.([]interface{})
		cur, _ = redis.Int(replyArray[0], nil)
		curKeys, _ := redis.Strings(replyArray[1], nil)
		var keys []string
		if nil != lastKeys {
			keys = append(lastKeys, curKeys...)
		} else {
			keys = curKeys
		}
		if len(keys) > maxLen {
			return nil, nerror.NewRunTimeError(fmt.Sprintf("Key数量超过了%d", maxLen))
		}
		if cur != 0 {
			return scanKeysWithConn(conn, cur, keyPattner, keys, maxLen)
		}
		return keys, nil
	}
	return nil, err
}

// expiry 过期时间-秒
// tries 重试次数
// delay 重试间隔时间
func (service *RedisService) LockRun(key, value string, expiry int, tries, delay int, runFun func() any) (result any, err error) {
	op1 := RedisMutexSetExpiry(time.Duration(expiry) * time.Second)
	op2 := RedisMutexSetDelay(time.Duration(delay) * time.Millisecond)
	op3 := RedisMutexSetTries(tries)
	mutex := RedisNewMutex(key, value, service, op1, op2, op3)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
		return runFun(), err
	}
	return nil, nerror.NewRunTimeError(fmt.Sprintf("[%v]-[%v]未获取到锁", key, value))
}

// 队列消息写入
func (service *RedisService) Producer(queueKey string, message string) error {
	conn := service.RedisPool.Get()
	defer conn.Close()
	replay, err := conn.Do("RPUSH", queueKey, message) // 或 "RPUSH"
	slog.Debug(fmt.Sprintf("队列[%s]中目前有[%v]条数据,当前写入[%v]", queueKey, replay, message))
	return err
}

// 队列消息读取
func (service *RedisService) Consumer(queueKey string, msgch chan string) {
	conn := service.RedisPool.Get()
	for {
		// BLPOP 返回格式: [队列名, 元素值]
		reply, err := redis.Strings(conn.Do("BLPOP", queueKey, 0)) // 0 表示无限阻塞
		if err != nil {
			slog.Warn(fmt.Sprintf("消费失败: %v, 3秒后重试中", err))
			//发生错误关闭
			conn.Close()
			time.Sleep(3 * time.Second)
			//重新连接
			conn = service.RedisPool.Get()
			continue
		} else {
			message := reply[1]
			msgch <- message
		}
	}
}
