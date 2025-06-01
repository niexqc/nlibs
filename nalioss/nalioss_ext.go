package nalioss

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/panjf2000/ants/v2"
)

type NAliOssClient struct {
	Cnf                     *nyaml.YamlConfNAliOssConf
	OssClient               *oss.Client
	MultipartUploadWorkPool *ants.Pool
}

func NewNAliOssClient(cnf *nyaml.YamlConfNAliOssConf) (*NAliOssClient, error) {
	var cfg = oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cnf.OssKey, cnf.OssKeySecret)).
		WithRegion("cn-chengdu").
		WithUseInternalEndpoint(cnf.InternalEndpoint)
	if cnf.ProxyEnabel {
		slog.Info(fmt.Sprintf("OSS当前为代理模式,通过代理【%v】访问", cnf.ProxyHttpUrl))
		cfg.WithProxyHost(cnf.ProxyHttpUrl)
	}
	wpool, err := ants.NewPool(cnf.MultipartUploadWorkNum, ants.WithNonblocking(false))
	if nil != err {
		return nil, nerror.NewRunTimeError("创建分片上传工作协程池失败")
	}
	return &NAliOssClient{Cnf: cnf, OssClient: oss.NewClient(cfg), MultipartUploadWorkPool: wpool}, nil
}

// ListObjects
func (svc *NAliOssClient) ListObjects(prefix string, pageSize int, pageCall func(pageNo int, data []oss.ObjectProperties)) error {
	// 创建列出对象的请求
	request := &oss.ListObjectsV2Request{
		Bucket:  oss.Ptr(svc.Cnf.BucketName),
		MaxKeys: int32(pageSize), //每次列举返回的最大对象数量
		Prefix:  oss.Ptr(prefix), // 列举指定前缀的所有对象
	}
	// 创建分页器
	paginator := svc.OssClient.NewListObjectsV2Paginator(request)
	pageNo := 0
	for paginator.HasNext() {
		pageNo++
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			slog.Error(fmt.Sprintf("分页获取前缀[%s]数据失败:%v", prefix, err))
			return err
		}
		pageCall(pageNo, page.Contents)
	}
	return nil
}

// UploadFile
func (svc *NAliOssClient) UploadFile(objKey, localFile string) error {
	putRequest := &oss.PutObjectRequest{
		Bucket: oss.Ptr(svc.Cnf.BucketName), // 存储空间名称
		Key:    oss.Ptr(objKey),             // 对象名称
	}
	_, err := svc.OssClient.PutObjectFromFile(context.TODO(), putRequest, localFile)
	if nil != err {
		slog.Error("文件上传到OSS失败:" + nerror.GenErrDetail(err))
		return err
	}
	return nil
}

// GetObj
func (svc *NAliOssClient) GetObj(objKey string) (result *oss.GetObjectResult, err error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(svc.Cnf.BucketName), // 存储空间名称
		Key:    oss.Ptr(objKey),             // 对象名称
	}
	return svc.OssClient.GetObject(context.TODO(), request)
}

// DeleteObj
func (svc *NAliOssClient) DeleteObj(objKey string) error {
	request := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(svc.Cnf.BucketName), // 存储空间名称
		Key:    oss.Ptr(objKey),             // 对象名称
	}
	// 执行删除对象的操作并处理结果
	_, err := svc.OssClient.DeleteObject(context.TODO(), request)
	if err != nil {
		slog.Error("文件删除失败:" + nerror.GenErrDetail(err))
		return err
	}
	return nil
}

// RunOssMultipartUpload  分片上传
func (svc *NAliOssClient) MultipartUpload(objKey, localFile string, chunkSize int64) error {
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return err
	}
	if fileInfo.Size() <= chunkSize {
		return svc.UploadFile(objKey, localFile)
	}
	count := (fileInfo.Size() / chunkSize)
	if (fileInfo.Size() % chunkSize) > 0 {
		count = count + 1
	}
	slog.Info(fmt.Sprintf("【%s】的大小为【%s】将被拆分为%d个分片上传:", localFile, ntools.FileSize2Str(fileInfo.Size()), count))
	// 初始化分片上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(svc.Cnf.BucketName),
		Key:    oss.Ptr(objKey),
	}
	initResult, err := svc.OssClient.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		slog.Error("初始化分片上传请求失败:" + nerror.GenErrDetail(err))
		return err
	}
	uploadId := *initResult.UploadId
	file, _ := os.Open(localFile)
	defer file.Close()

	// 初始化等待组和互斥锁
	partNumber := int64(0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	parts := make([]oss.UploadPart, 0)
	for {
		offset := partNumber * chunkSize
		currentChunkSize := min(chunkSize, fileInfo.Size()-offset)
		if currentChunkSize <= 0 {
			break
		}
		chunkData := make([]byte, currentChunkSize)
		file.Read(chunkData)

		wg.Add(1)
		curPartNumber := partNumber + 1
		slog.Debug(fmt.Sprintf("分片序号:%d,当前分片大小:%s", curPartNumber, ntools.FileSize2Str(currentChunkSize)))

		svc.MultipartUploadWorkPool.Submit(func() {
			// 创建分片上传请求
			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(svc.Cnf.BucketName), // 目标存储空间名称
				Key:        oss.Ptr(objKey),             // 目标对象名称
				PartNumber: int32(curPartNumber),        // 分片编号
				UploadId:   oss.Ptr(uploadId),           // 上传ID
				Body:       bytes.NewReader(chunkData),  // 分片内容
			}
			// 发送分片上传请求
			partResult, err := retryUploadPart(svc.OssClient, partRequest, 1, 3)
			if err != nil {
				slog.Error(fmt.Sprintf("分片上传失败 %d: %v", curPartNumber, err))
			}
			// 记录分片上传结果
			mu.Lock()
			parts = append(parts, oss.UploadPart{PartNumber: partRequest.PartNumber, ETag: partResult.ETag})
			mu.Unlock()
			wg.Done()
			slog.Debug(fmt.Sprintf("分片序号:%d,已上传完成", curPartNumber))
		})
		// 增加
		partNumber++
	}
	wg.Wait()
	// 完成分片上传请求
	request := &oss.CompleteMultipartUploadRequest{
		Bucket:                  oss.Ptr(svc.Cnf.BucketName),
		Key:                     oss.Ptr(objKey),
		UploadId:                oss.Ptr(uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{Parts: parts},
	}
	_, err = svc.OssClient.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		slog.Error("完成分片上传请求，执行失败:" + nerror.GenErrDetail(err))
		return err
	}
	slog.Info(fmt.Sprintf("本地文件:%s,已上传到:%s", localFile, objKey))
	return err
}

func retryUploadPart(client *oss.Client, requst *oss.UploadPartRequest, curTimes, retryMaxTimes int) (partResult *oss.UploadPartResult, err error) {
	slog.Debug(fmt.Sprintf("分片序号:%v,第%v/%v次上传", requst.PartNumber, curTimes, retryMaxTimes))
	partResult, err = client.UploadPart(context.TODO(), requst)
	curTimes = curTimes + 1
	if err == nil || curTimes > retryMaxTimes {
		return partResult, nil
	}
	return retryUploadPart(client, requst, curTimes, retryMaxTimes)
}
