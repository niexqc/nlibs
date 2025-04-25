package ntools

import (
	"bufio"
	"container/list"
	"errors"

	"os"
	"path"
	"path/filepath"
	"strings"
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

// PathDir 获取文件的目录
func (fde *fileDirExt) PathDir(filePath string) string {
	return strings.TrimSuffix(filePath, path.Base(filePath))
}

// PathFileSuffix 获取文件后缀名
func (fde *fileDirExt) PathFileSuffix(filePath string) string {
	return path.Ext(filePath)
}

// PathFileName 获取文件名字不包含后缀
func (fde *fileDirExt) PathFileName(filePath string) string {
	return strings.TrimSuffix(fde.PathFileNameWithSuffix(filePath),
		fde.PathFileSuffix(filePath))
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

// WriteFileContent 写入文件内容，目录|文件不存在则创建目录|文件
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) WriteFileContent(filename string, content string, append bool) (bool, error) {
	return fde.WriteFileByWriterFun(filename, func(outputWriter *bufio.Writer) {
		outputWriter.WriteString(content)
	}, append)
}

// WriteFile 写入文件内容，目录|文件不存在则创建目录|文件
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) WriteFile(filename string, content *[]byte, append bool) (bool, error) {
	return fde.WriteFileByWriterFun(filename, func(outputWriter *bufio.Writer) {
		outputWriter.Write(*content)
	}, append)
}

// WriteFileByWriterFun 写入文件内容，目录|文件不存在则创建目录|文件
//
//	Return  存在返回 true 不存在返回false
func (fde *fileDirExt) WriteFileByWriterFun(filename string, writeFun func(*bufio.Writer), append bool) (bool, error) {
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
		return nil, errors.New("文件不存在")
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

// TraverseDir 递归文件夹获取到所有文件名称
//
//	dirPth 目录
func (fde *fileDirExt) TraverseDir(dirPth string, fileList *list.List) error {
	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return err
	}
	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			fde.TraverseDir(dirPth+PthSep+fi.Name(), fileList)
		} else {
			fileList.PushBack(dirPth + PthSep + fi.Name())
		}
	}
	return nil
}

// TraverseDirBySlice 递归文件夹获取到所有文件名称
//
//	dirPth 目录
func (fde *fileDirExt) TraverseDirBySlice(dirPth string) ([]string, error) {
	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)

	var curFile []string
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			down, _ := fde.TraverseDirBySlice(dirPth + PthSep + fi.Name())
			curFile = append(curFile, down...)
		} else {
			curFile = append(curFile, dirPth+PthSep+fi.Name())
		}
	}

	return curFile, nil
}
