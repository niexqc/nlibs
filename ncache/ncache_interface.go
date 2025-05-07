package ncache

// 谨慎使用memcache
//
type INcache interface {
	// Int64自增
	Int64Incr(key string, expireMillisecond int64) (num int64, err error)
	// PutStr
	PutStr(key string, val string) error
	// 无论key是否存在，并指定过期时间（​​原子性操作​​），都会​​覆盖旧值​​并设置新的过期时间
	PutExStr(key string, val string, sencond int) error
	// 仅在【key​不存在】​​时成功（​​原子性操作​​）
	PutNxExStr(key string, val string, sencond int) error
	// GetStr
	GetStr(key string) (string, error)
	// 是否存在某个key
	Exist(key string) (bool, error)
	// 是否存在某个Key
	ExistWithoutErr(key string) bool
	// 对某个值设置过期时间
	KeySetExpire(key string, sencond int) error
	// ClearByKeyPrefix
	ClearByKeyPrefix(keyPrefix string) (int, error)
	//  ClearKey
	ClearKey(key string) error
}
