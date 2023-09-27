package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func RunCmd(ctx context.Context, cmd ...string) (int, []byte, error) {
	var c *exec.Cmd

	if ctx == nil {
		err := exec.Command("sh", "-c", strings.Join(cmd, " ")).Run()
		return 0, nil, err
	}

	c = exec.CommandContext(ctx, "sh", "-c", strings.Join(cmd, " "))
	sysProcAttr(c)

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout // 标准输出
	c.Stderr = &stderr // 标准错误
	output := make([]byte, 0)
	err := c.Run()
	if len(stderr.Bytes()) > 0 {
		output = append(output, stderr.Bytes()...)
	}
	if len(stdout.Bytes()) > 0 {
		output = append(output, stdout.Bytes()...)
	}

	return c.Process.Pid, output, err
}

// StartApp
// @load/cb 启动启动程序后load秒执行脚本或函数（例如加载配置等）
// #cmd[0] 为App的路径
func StartApp(ctx context.Context, l *logrus.Logger, tag string, load int, cb func(context.Context), cmd ...string) {
	for {
		l.Info(fmt.Sprintf("%s:启动", tag))
		//App启动后通过回调函数cb,加载策略
		ctxCb, cancelCb := context.WithCancel(ctx)
		if cb != nil {
			go func() {
				tk := time.NewTicker(time.Second * time.Duration(load))
				defer tk.Stop()
				select {
				case <-tk.C: //等待一个load周期加载策略
					cb(ctxCb)
				case <-ctxCb.Done():
					return
				}
			}()
		}
		//杀死fixme: pid @ $1
		_, _, err := RunCmd(ctx,
			fmt.Sprintf("kill -9 `ps aux|grep %s|grep -v grep|awk '{print $1}'`", tag))
		if err != nil {
			l.Debug(fmt.Sprintf("Stop: %s: %s failed", tag, err.Error()))
		}
		//前台运行App
		pid, _, err := RunCmd(ctx, cmd...)
		if err != nil {
			l.Error(fmt.Sprintf("StartApp: %s: %s (will restart after %d seconds)", tag, err.Error(), load))
		}
		cancelCb()
		l.Info(fmt.Sprintf("%s[%d]:退出", tag, pid))
		if load == 0 {
			break
		}
		time.Sleep(time.Second * time.Duration(load))
		//如果App结束，应取消策略回调
	}
}
