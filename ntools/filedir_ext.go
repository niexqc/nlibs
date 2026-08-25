package ntools

import (
	"bufio"

	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/niexqc/nlibs/nerror"
)

type fileDirExt struct{}

var _fileDirExt = &fileDirExt{}

func GetFileDirExt() *fileDirExt {
	return _fileDirExt
}

// JoinPath 拼接路径
func (fde *fileDirExt) JoinPath(items ...string) string {
	return path.Join(items...)
}

// PathFileSuffix 获取文件后缀名
func (fde *fileDirExt) PathFileSuffix(filePath string) string {
	return path.Ext(filePath)
}

// PathFileNameWithSuffix 获取文件名字包含后缀
func (fde *fileDirExt) PathFileNameWithSuffix(filePath string) string {
	return path.Base(filePath)
}

// CheckFileIsExist 判断文件是否存在
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// MkDirIfNotExist 如果目录不存在则创建目录
func (fde *fileDirExt) MkAllDirIfNotExist(dir string) error {
	exist := fde.CheckFileIsExist(dir)
	if !exist {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// WriteFileContent 写入文件内容，目录|文件不存在则创建目录|文件
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) WriteFileContent(filename string, content string, append bool) (bool, error) {
	return fde.writeFileByWriterFun(filename, func(outputWriter *bufio.Writer) {
		outputWriter.WriteString(content)
	}, append)
}

// WriteFile 写入文件内容，目录|文件不存在则创建目录|文件
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) WriteFile(filename string, content *[]byte, append bool) (bool, error) {
	return fde.writeFileByWriterFun(filename, func(outputWriter *bufio.Writer) {
		outputWriter.Write(*content)
	}, append)
}

func (fde *fileDirExt) writeFileByWriterFun(filename string, writeFun func(*bufio.Writer), append bool) (bool, error) {
	dir := filepath.Dir(filename)
	if !fde.CheckFileIsExist(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}
	//如果不是追加模式，则删除旧文件再写入
	if !append {
		os.Remove(filename)
	}
	var flag int
	if append {
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	} else {
		flag = os.O_RDWR | os.O_CREATE
	}
	outputFile, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return false, err
	}
	defer outputFile.Close()
	outputWriter := bufio.NewWriter(outputFile)
	defer outputWriter.Flush()
	//写入内容
	writeFun(outputWriter)
	return true, nil
}

// ReadFileByte 读取文本文件内容
func (fde *fileDirExt) ReadFileByte(filename string) ([]byte, error) {
	if !fde.CheckFileIsExist(filename) {
		return nil, nerror.NewRunTimeErrorFmt("文件%s不存在", filename)
	}
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadFileContent 读取文本文件内容
func (fde *fileDirExt) ReadFileContent(filename string) (string, error) {
	data, err := fde.ReadFileByte(filename)
	if nil != err {
		return "", err
	}
	return string(data), nil
}

// 获取直接子目录
func (fde *fileDirExt) GetDirectSubDirs(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath) // 读取目录条目
	if err != nil {
		return nil, err
	}

	var subDirs []string
	for _, entry := range entries {
		if entry.IsDir() { // 判断是否为目录
			subDirs = append(subDirs, filepath.Join(dirPath, entry.Name()))
		}
	}
	return subDirs, nil
}

// dirSize 返回指定目录的总大小（以字节为单位）
func (fde *fileDirExt) DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// PathCutPreFix 去除路径前缀
func (fde *fileDirExt) PathCutPreFix(fullPathStr, prefixPath string) string {
	fullPathStr = strings.TrimSuffix(strings.ReplaceAll(fullPathStr, "\\", "/"), "/")
	prefixPath = strings.TrimPrefix(strings.ReplaceAll(prefixPath, "\\", "/"), "/")
	resultStr := strings.TrimPrefix(fullPathStr, prefixPath)
	resultStr = strings.TrimPrefix(resultStr, "/")
	return resultStr
}
