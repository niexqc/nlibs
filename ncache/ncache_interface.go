package ncache

type INcache interface {
	Int64Incr(key string, expireMillisecond int64) (num int64, err error)
	PutStr(key string, val string) error
	GetStr(key string) (string, error)
	ExistWithoutErr(key string) bool
	ExpireKey(key string, sencond int) error
	ClearByKeyPrefix(keyPrefix string) (int, error)
	PutExStr(key string, val string, sencond int) error
	ClearKey(key string) error
}
