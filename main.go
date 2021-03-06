package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/constant"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub"
	"github.com/Dreamacro/clash/hub/executor"
	"github.com/Dreamacro/clash/log"
)

var (
	flagset            map[string]bool
	version            bool
	testConfig         bool
	homeDir            string
	configFile         string
	externalUI         string
	externalController string
	secret             string
)

func init() {
	flag.StringVar(&homeDir, "d", "", "set configuration directory")
	flag.StringVar(&configFile, "f", "", "specify configuration file")
	flag.StringVar(&externalUI, "ext-ui", "", "override external ui directory")
	flag.StringVar(&externalController, "ext-ctl", "", "override external controller address")
	flag.StringVar(&secret, "secret", "", "override secret for RESTful API")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.BoolVar(&testConfig, "t", false, "test configuration and exit")
	flag.Parse()

	// 短选项开关
	flagset = map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		flagset[f.Name] = true
	})
}

func main() {
	if version {
		fmt.Printf("Clash %s %s %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, C.BuildTime)
		return
	}

	//获取homDir的绝对路径
	if homeDir != "" {
		if !filepath.IsAbs(homeDir) {
			currentDir, _ := os.Getwd()
			log.Infoln("..currentDir...%s", currentDir)
			homeDir = filepath.Join(currentDir, homeDir)
			log.Infoln("..homeDir...%s", homeDir)
		}
		C.SetHomeDir(homeDir)
	}

	// 设置配置文件
	if configFile != "" {
		if !filepath.IsAbs(configFile) {
			// 获取当前目录
			currentDir, _ := os.Getwd()
			// 根据当前路径来产生绝对路径
			configFile = filepath.Join(currentDir, configFile)
		}
		C.SetConfig(configFile)
	} else {
		// 设置默认的文件路径
		configFile := filepath.Join(C.Path.HomeDir(), C.Path.Config())
		log.Infoln("..configFile...%s", configFile)
		C.SetConfig(configFile)
	}

	// 创建config.yaml和下载mmdb
	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}

	if testConfig {
		if _, err := executor.Parse(); err != nil {
			log.Errorln(err.Error())
			fmt.Printf("configuration file %s test failed\n", constant.Path.Config())
			os.Exit(1)
		}
		fmt.Printf("configuration file %s test is successful\n", constant.Path.Config())
		return
	}

	var options []hub.Option
	if flagset["ext-ui"] {
		log.Infoln("has ext-uti")
		options = append(options, hub.WithExternalUI(externalUI))
	}
	if flagset["ext-ctl"] {
		log.Infoln("has ext-ctl")
		options = append(options, hub.WithExternalController(externalController))
	}
	if flagset["secret"] {
		log.Infoln("has secret")
		options = append(options, hub.WithSecret(secret))
	}

	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	sigCh := make(chan os.Signal, 1)
	// 接收SIGTERM信号
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
