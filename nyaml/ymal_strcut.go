package nyaml

type YamlConfLog struct {
	LogNamePrefix string `yaml:"logNamePrefix" hc:"日志文件的前缀"`
	LogLevel      string `yaml:"logLevel" hc:" debug|info|warn|error "`
	OutMode       int    `yaml:"outMode" hc:" 日志输出方式: 0-不打印,1-控制台,2-文件,3-都打印 "`
	PrintMethod   int    `yaml:"printMethod" hc:" 打印日志发生的详情: 0-不打印,1-详情,2-仅方法名称"`
}

type YamlConfRedis struct {
	RedisHost      string `yaml:"redisHost" hc:"redisHost"`
	RedisPort      int    `yaml:"redisPort" hc:"redisPort"`
	RedisPwd       string `yaml:"redisPwd" hc:"redisPwd"`
	DataBaseIdx    int    `yaml:"dataBaseIdx" hc:"dataBaseIdx"`
	ConnectTimeout int    `yaml:"connectTimeout" hc:"连接超时时间-秒"`
	ReadTimeout    int    `yaml:"readTimeout" hc:"读取超时时间-秒"`
	WriteTimeout   int    `yaml:"writeTimeout" hc:"写入超时时间-秒"`
	MaxIdle        int    `yaml:"maxIdle" hc:"MaxIdle"`
	MaxActive      int    `yaml:"maxActive" hc:"MaxActive"`
	IdleTimeout    int    `yaml:"idleTimeout" hc:"IdleTimeout-秒"`
}

type YamlConfSqlPrint struct {
	DbSqlLogPrint    bool   `yaml:"dbSqlLogPrint" hc:"Sql日志是否打印 true|false"`
	DbSqlLogLevel    string `yaml:"dbSqlLogLevel" hc:"Sql日志使用【 debug|info|warn|error 】输出"`
	DbSqlLogCompress bool   `yaml:"dbSqlLogCompress" hc:"Sql日志打印是否压缩 true|false"`
}

// 历史原因这个用于mysql
type YamlConfMysqlDb struct {
	DbHost          string `yaml:"dbHost" hc:"dbHost"`
	DbPort          int64  `yaml:"dbPort" hc:"dbPort"`
	DbUser          string `yaml:"dbUser" hc:"dbUser"`
	DbPwd           string `yaml:"dbPwd" hc:"dbPwd"`
	DbName          string `yaml:"dbName" hc:"DbName"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime" hc:"连接最大时长-秒"`
	MaxOpenConns    int    `yaml:"maxOpenConns" hc:"MaxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns" hc:"MaxIdleConns"`
}

type YamlConfPgDb struct {
	DbHost          string `yaml:"dbHost" hc:"dbHost"`
	DbPort          int64  `yaml:"dbPort" hc:"dbPort"`
	DbUser          string `yaml:"dbUser" hc:"dbUser"`
	DbPwd           string `yaml:"dbPwd" hc:"dbPwd"`
	DbName          string `yaml:"dbName" hc:"DbName"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime" hc:"连接最大时长-秒"`
	MaxOpenConns    int    `yaml:"maxOpenConns" hc:"MaxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns" hc:"MaxIdleConns"`
}

type YamlConfEndnKey struct {
	Sm2HexPubKey string `yaml:"sm2HexPubKey" hc:"服务端SM2公钥"`
	Sm2HexPriKey string `yaml:"sm2HexPriKey" hc:"服务端SM2私钥"`
}

type YamlConfNAliOssConf struct {
	InternalEndpoint       bool   `yaml:"internalEndpoint" hc:"程序是否运行在OSS所在地域内网"`
	BucketName             string `yaml:"bucketName" hc:"Bucket名称"`
	OssRegion              string `yaml:"ossRegion" hc:"Bucket所在地区的地址[cn-chengdu]"`
	OssKey                 string `yaml:"ossKey" hc:"OssKey"`
	OssKeySecret           string `yaml:"ossKeySecret" hc:"OssKeySecret"`
	OssPrefix              string `yaml:"ossPrefix" hc:"Oss存储的前缀"`
	ProxyEnabel            bool   `yaml:"proxyEnabel" hc:"是否开启代理"`
	ProxyHttpUrl           string `yaml:"proxyHttpUrl" hc:"代理的地址"`
	MultipartUploadWorkNum int    `yaml:"multipartUploadWorkNum" hc:"分片上传文件最大的并发数"`
}
