package nalioss_test

import (
	"fmt"
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
	aliOssClient = nalioss.NewNAliOssClient(aliOssConf)
}

func TestListObjects(t *testing.T) {
	aliOssClient.ListObjects(aliOssConf.OssPrefix, 10, func(pageNo int, data []oss.ObjectProperties) {
		fmt.Println(pageNo, data)
	})
}

func TestUploadFile(t *testing.T) {
	localFile := "nalioss_ext_test.go"
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, "nalioss_ext_test.go")
	err := aliOssClient.UploadFile(objKey, localFile)
	if nil != err {
		t.Errorf("TestUploadFile:%v", nerror.GenErrDetail(err))
	}
}

func TestDeleteObj(t *testing.T) {
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, "nalioss_ext_test.go")
	err := aliOssClient.DeleteObj(objKey)
	if nil != err {
		t.Errorf("TestDeleteObj:%v", nerror.GenErrDetail(err))
	}
}

func TestMultipartUpload(t *testing.T) {
	objKey := nlibs.FileDirExt.JoinPath(aliOssConf.OssPrefix, "test.aa")
	err := aliOssClient.MultipartUpload(objKey, "/bigfile/aa.msi", 1*1024*1024)
	if nil != err {
		t.Errorf("TestDeleteObj:%v", nerror.GenErrDetail(err))
	}
}
