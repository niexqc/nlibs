package ntools

import (
	"bufio"
	"bytes"

	"io"
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
	if err := fde.MkAllDirIfNotExist(filepath.Dir(filename)); err != nil {
		return false, err
	}
	if append {
		return fde.appendFileByWriterFun(filename, writeFun)
	}
	//覆盖写入先把内容渲染到内存，便于整体替换或重写
	contentBuffer := new(bytes.Buffer)
	if err := writeAndFlushBufio(contentBuffer, writeFun); err != nil {
		return false, err
	}
	return fde.overwriteFileByWriterFun(filename, contentBuffer.Bytes())
}

// appendFileByWriterFun 追加写入
func (fde *fileDirExt) appendFileByWriterFun(filename string, writeFun func(*bufio.Writer)) (bool, error) {
	outputFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return false, err
	}
	writeErr := writeAndFlushBufio(outputFile, writeFun)
	if closeErr := outputFile.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		return false, writeErr
	}
	return true, nil
}

// overwriteFileByWriterFun 覆盖写入
//
// 先写同目录下的临时文件，刷盘成功后再原子替换目标文件；
// 目标文件被占用等原因导致替换失败时，退化成“截断后原地重写”。
//
// 旧实现是【os.Remove(删除失败的错误被忽略) + O_RDWR(不带 O_TRUNC)】：
// 一旦旧文件删除失败(如docker单文件挂载、杀毒软件占用)，
// 就只会从偏移0处覆盖写入，旧内容的尾部会残留在文件中形成乱码。
func (fde *fileDirExt) overwriteFileByWriterFun(filename string, content []byte) (bool, error) {
	if err := fde.replaceFileByRename(filename, content); err == nil {
		return true, nil
	}
	//替换失败时退化为原地重写，O_TRUNC保证不会残留旧内容
	return fde.truncateAndWriteFile(filename, content)
}

// replaceFileByRename 通过【临时文件+rename】原子替换目标文件
func (fde *fileDirExt) replaceFileByRename(filename string, content []byte) error {
	if err := fde.MkAllDirIfNotExist(filepath.Dir(filename)); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp*")
	if err != nil {
		return err
	}
	tmpFileName := tmpFile.Name()
	finished := false
	//失败时清理临时文件，避免残留
	defer func() {
		if !finished {
			tmpFile.Close()
			os.Remove(tmpFileName)
		}
	}()

	if err := tmpFile.Chmod(fde.fileModeOrDefault(filename)); err != nil {
		return err
	}
	if _, err := tmpFile.Write(content); err != nil {
		return err
	}
	//先落盘，避免替换成功后内容还未真正写入磁盘
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpFileName, filename); err != nil {
		return err
	}
	finished = true
	return nil
}

// truncateAndWriteFile 截断目标文件后原地重写
func (fde *fileDirExt) truncateAndWriteFile(filename string, content []byte) (bool, error) {
	//O_TRUNC保证旧内容不会残留在文件中
	outputFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return false, err
	}
	writeErr := writeAndFlushBufio(outputFile, func(outputWriter *bufio.Writer) {
		outputWriter.Write(content)
	})
	if closeErr := outputFile.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		return false, writeErr
	}
	return true, nil
}

// fileModeOrDefault 目标文件已存在时沿用其权限，否则使用0666
func (fde *fileDirExt) fileModeOrDefault(filename string) os.FileMode {
	if fileInfo, err := os.Stat(filename); err == nil {
		return fileInfo.Mode().Perm()
	}
	return 0666
}

// writeAndFlushBufio 执行写入并刷出缓冲区，不再忽略Flush的错误
func writeAndFlushBufio(outputFile io.Writer, writeFun func(*bufio.Writer)) error {
	outputWriter := bufio.NewWriter(outputFile)
	writeFun(outputWriter)
	return outputWriter.Flush()
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
