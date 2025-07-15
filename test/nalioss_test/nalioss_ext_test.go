package nalioss_test

import (
	"crypto/rand"
	"log/slog"
	"os"
	"testing"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/nalioss"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var aliOssClient *nalioss.NAliOssClient
var aliOssConf *nyaml.YamlConfNAliOssConf

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	aliOssConf = &nyaml.YamlConfNAliOssConf{
		InternalEndpoint:       false,
		BucketName:             "wts-backup-sanzi2025",
		OssRegion:              "cn-chengdu",
		OssKey:                 "",
		OssKeySecret:           "",
		OssPrefix:              "xbak/dev253",
		ProxyEnabel:            false,
		ProxyHttpUrl:           "http://192.168.0.251:1080",
		MultipartUploadWorkNum: 30,
	}
	aliOssConf.OssKey = os.Getenv("WTS_TEST_OSS_Key")
	aliOssConf.OssKeySecret = os.Getenv("WTS_TEST_OSS_Secret")
	if aliOssConf.OssKey == "" || aliOssConf.OssKeySecret == "" {
		panic(nerror.NewRunTimeError("请在环境变量配置:WTS_TEST_OSS_Key,WTS_TEST_OSS_Secret"))
	}
	var err error
	aliOssClient, err = nalioss.NewNAliOssClient(aliOssConf)
	if nil != err {
		panic(err)
	}
}

func TestListObjects(t *testing.T) {

	currentPage := new(int)
	*currentPage = 0
	datas := []oss.ObjectProperties{}

	err := aliOssClient.ListObjects(aliOssConf.OssPrefix, 10, func(pageNo int, data []oss.ObjectProperties) {
		*currentPage = pageNo
		datas = append(datas, data...)
		slog.Info("TestListObjects", "pageNo", pageNo, " DataLen", len(data))
	})
	ntools.TestErrPainic(t, "TestListObjects", err)
	ntools.TestEq(t, "TestListObjects pageNo 必须大于0", true, *currentPage > 0)
	ntools.TestEq(t, "TestListObjects data_len 必须大于0", true, len(datas) > 0)
}

func TestUploadFile(t *testing.T) {
	localFile := "nalioss_ext_test.go"
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, localFile)

	err := aliOssClient.UploadFile(objKey, localFile)
	ntools.TestErrPainic(t, "TestUploadFile", err)

	aliOssClient.DeleteObj(objKey)

}

func TestDeleteObj(t *testing.T) {
	localFile := "nalioss_ext_test.go"
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, localFile)

	aliOssClient.UploadFile(objKey, localFile)
	err := aliOssClient.DeleteObj(objKey)
	ntools.TestErrPainic(t, "TestDeleteObj", err)
}

func TestMultipartUpload(t *testing.T) {
	localFile := "test.aa"
	//分片上传需要先生成一个临时文件
	for i := 0; i < 6; i++ {
		buffer := make([]byte, 1024*1024) // 1Mb
		rand.Read(buffer)
		nlibs.FileDirExt.WriteFile(localFile, &buffer, true)
	}
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, localFile)
	err := aliOssClient.MultipartUpload(objKey, localFile, 1*1024*1024)
	ntools.TestErrPainic(t, "TestMultipartUpload", err)

	//删除临时文件
	os.Remove(localFile)
	//清理上传的文件
	aliOssClient.DeleteObj(objKey)
}
