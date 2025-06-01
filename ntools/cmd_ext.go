package ntools

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 执行命令，命令执行完成后,函数内部主动关闭（cmdOut）
func CmdRunAndPrintLog(windows bool, command, workDir string, args ...string) error {
	str, _ := filepath.Abs(workDir)
	slog.Debug("当前工作目录:" + str)
	cmdOut := make(chan string, 10)
	go func() {
		for v := range cmdOut {
			slog.Info(v)
		}
	}()
	return CmdRunWithStdOut(windows, command, workDir, cmdOut, args...)
}

// 执行命令，命令执行完成后,函数内部主动关闭（cmdOut）
func CmdRunWithStdOut(windows bool, command, workDir string, cmdOut chan string, arg ...string) error {
	str, _ := filepath.Abs(workDir)
	slog.Debug("当前工作目录:" + str)
	defer close(cmdOut)
	cmd := exec.Command(command, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error(fmt.Sprintf("cmd.StdoutPipe:%v ", err))
		return err
	}
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir

	err = cmd.Start()
	if err != nil {
		return err
	}
	//创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
	reader := bufio.NewReader(stdout)
	for {
		lineBytes, err2 := reader.ReadBytes(byte('\n'))
		if err2 != nil || io.EOF == err2 {
			break
		}
		if OsIsWindows() {
			linestr, _ := strings.CutSuffix(StrFromGbkBytes(lineBytes).S, "\r\n")
			cmdOut <- linestr
		} else {
			linestr, _ := strings.CutSuffix(string(lineBytes), "\n")
			cmdOut <- linestr
		}
	}
	return cmd.Wait()
}
