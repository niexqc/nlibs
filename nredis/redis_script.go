package nredis

import "github.com/gomodule/redigo/redis"

var s_delkey_value = `
if redis.call('GET', KEYS[1])==ARGV[1] then
	return redis.call('DEL', KEYS[1])
else
	return 0
end
`
var s_int_incr = `
local current = redis.call('incr',KEYS[1]);
local t = redis.call('ttl',KEYS[1]); 
if t == -1 then
	redis.call('pexpire',KEYS[1],ARGV[1])
end;
return current
`

var ScriptDelKv = redis.NewScript(1, s_delkey_value)
var ScriptIntIncr = redis.NewScript(1, s_int_incr)
