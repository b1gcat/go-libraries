package utils

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/sirupsen/logrus"
	//go tool pprof  https+insecure://localhost:8888/debug/pprof/profile?seconds=60
	//_ "net/http/pprof"
)

var fPprof *os.File

// 优雅退出
func waitExit(c chan os.Signal) {
	for i := range c {
		switch i {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			logrus.Info("生成运行报告 ", i.String(), ",exit...")

			// CPU 性能分析
			fPprof.Close()
			pprof.StopCPUProfile()

			os.Exit(0)
		}
	}
}

// Pprof go tool pprof vpnServer_xxx cpu.prof
func Pprof(name string) {
	logrus.Info("开启调试模式:", "go tool pprof "+name)
	//CPU 性能分析
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error("os.OpenFile:", err.Error())
		return
	}
	_ = pprof.StartCPUProfile(f)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		waitExit(c)
	}()
}
