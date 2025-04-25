package nyaml

type YamlConfLog struct {
	LogNamePrefix string `yaml:"logNamePrefix" hc:"日志文件的前缀"`
	LogLevel      string `yaml:"logLevel" hc:" debug|info|warn|error "`
	OutMode       int    `yaml:"stdOut" hc:" 日志输出方式: 0-不打印,1-控制台,2-文件,3-都打印 "`
	PrintMethod   int    `yaml:"printMethod" hc:" 打印日志发生的详情: 0-不打印,1-详情,2-仅方法名称"`
}

type YamlConfRedis struct {
	RedisHost string `yaml:"redisHost" hc:"redisHost"`
	RedisPort int    `yaml:"redisPort" hc:"redisPort"`
	RedisPwd  string `yaml:"redisPwd" hc:"redisPwd"`
}

type YamlConfDb struct {
	DbHost string `yaml:"dbHost" hc:"dbHost"`
	DbPort int64  `yaml:"dbPort" hc:"dbPort"`
	DbUser string `yaml:"dbUser" hc:"dbUser"`
	DbPwd  string `yaml:"dbPwd" hc:"dbPwd"`
	DbName string `yaml:"dbName" hc:"DbName"`

	DbSqlLogPrint    bool   `yaml:"dbSqlLogPrint" hc:"Sql日志是否打印 true|false"`
	DbSqlLogLevel    string `yaml:"dbSqlLogLevel" hc:"Sql日志使用【 debug|info|warn|error 】输出"`
	DbSqlLogCompress bool   `yaml:"dbLogCompress" hc:"Sql日志打印是否压缩 true|false"`
}

type YamlConfEndnKey struct {
	Sm2HexPubKey string `yaml:"sm2HexPubKey" hc:"服务端SM2公钥"`
	Sm2HexPriKey string `yaml:"sm2HexPriKey" hc:"服务端SM2私钥"`
}
