package nalioss

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/panjf2000/ants/v2"
)

// 经 HTTP 代理上传大分片时，SDK 默认连接/读写超时过短易触发 TLS handshake timeout、write i/o timeout。
const ossConnectTimeout = 60 * time.Second
const ossReadWriteTimeout = 15 * time.Minute

// 单 HTTP 代理同时承载过多 UploadPart（大 body）易排队超时，启用代理时对并发做上限。
const multipartWorkersBehindProxy = 10

type NAliOssClient struct {
	Cnf                     *nyaml.YamlConfNAliOssConf
	OssClient               *oss.Client
	MultipartUploadWorkPool *ants.Pool
}

func NewNAliOssClient(cnf *nyaml.YamlConfNAliOssConf) (*NAliOssClient, error) {
	var cfg = oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cnf.OssKey, cnf.OssKeySecret)).
		WithRegion("cn-chengdu").
		WithUseInternalEndpoint(cnf.InternalEndpoint).
		WithConnectTimeout(ossConnectTimeout).
		WithReadWriteTimeout(ossReadWriteTimeout)
	// 是否开启代理
	if cnf.ProxyEnable {
		slog.Info(fmt.Sprintf("OSS当前为代理模式,通过代理【%v】访问", cnf.ProxyHttpUrl))
		cfg.WithProxyHost(cnf.ProxyHttpUrl)
	}
	// 分片上传文件最大的并发数
	workNum := cnf.MultipartUploadWorkNum
	if workNum < 1 {
		workNum = 4
	}
	// 如果开启代理，并且分片上传文件最大的并发数大于10，则调整为10
	if cnf.ProxyEnable && workNum > multipartWorkersBehindProxy {
		slog.Info(fmt.Sprintf("OSS 经代理上传: 分片并发由 %d 调整为 %d，降低代理侧超时风险", cnf.MultipartUploadWorkNum, multipartWorkersBehindProxy))
		workNum = multipartWorkersBehindProxy
	}

	wpool, err := ants.NewPool(workNum, ants.WithNonblocking(false))
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
			// 发送分片上传请求，最多尝试 5 次（含短暂退避，缓解代理瞬时拥塞）
			partResult, err2 := retryUploadPart(svc.OssClient, partRequest, 1, 5)
			if err2 != nil {
				slog.Error(fmt.Sprintf("分片上传失败 %d: %v", curPartNumber, err2))
			} else {
				// 记录分片上传结果
				mu.Lock()
				parts = append(parts, oss.UploadPart{PartNumber: partRequest.PartNumber, ETag: partResult.ETag})
				mu.Unlock()
				slog.Debug(fmt.Sprintf("分片序号:%d,已上传完成", curPartNumber))
			}
			wg.Done()
		})
		// 增加
		partNumber++
	}
	wg.Wait()
	if int64(len(parts)) != count {
		return fmt.Errorf("分片上传未全部成功: 期望 %d 片, 实际完成 %d 片", count, len(parts))
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})
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

func retryUploadPart(client *oss.Client, request *oss.UploadPartRequest, attempt, maxAttempts int) (*oss.UploadPartResult, error) {
	slog.Debug(fmt.Sprintf("分片序号:%v,第%v/%v次上传", request.PartNumber, attempt, maxAttempts))
	partResult, err := client.UploadPart(context.TODO(), request)
	if err == nil {
		return partResult, nil
	}
	slog.Error(fmt.Sprintf("分片序号:%v,第%v/%v次上传失败:%v", request.PartNumber, attempt, maxAttempts, err))
	if attempt >= maxAttempts {
		return nil, err
	}
	time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	return retryUploadPart(client, request, attempt+1, maxAttempts)
}
