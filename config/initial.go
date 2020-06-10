package config

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Dreamacro/clash/component/mmdb"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
)

func downloadMMDB(path string) (err error) {
	// 下载Country.mmdb
	resp, err := http.Get("https://github.com/Dreamacro/maxmind-geoip/releases/latest/download/Country.mmdb")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 创建相关文件
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	// 向该文件写入
	_, err = io.Copy(f, resp.Body)

	return err
}

// 初始化
func initMMDB() error {
	if _, err := os.Stat(C.Path.MMDB()); os.IsNotExist(err) {
		log.Infoln("Can't find MMDB, start download")
		if err := downloadMMDB(C.Path.MMDB()); err != nil {
			return fmt.Errorf("Can't download MMDB: %s", err.Error())
		}
	}

	if !mmdb.Verify() {
		log.Warnln("MMDB invalid, remove and download")
		if err := os.Remove(C.Path.MMDB()); err != nil {
			return fmt.Errorf("Can't remove invalid MMDB: %s", err.Error())
		}

		// 下载全球ip库
		if err := downloadMMDB(C.Path.MMDB()); err != nil {
			return fmt.Errorf("Can't download MMDB: %s", err.Error())
		}
	}

	return nil
}

// Init prepare necessary files
func Init(dir string) error {
	// initial homedir
	// 创建主目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("Can't create config directory %s: %s", dir, err.Error())
		}
	}

	// initial config.yaml
	if _, err := os.Stat(C.Path.Config()); os.IsNotExist(err) {
		log.Infoln("Can't find config, create a initial config file")
		// 创建config.yaml
		f, err := os.OpenFile(C.Path.Config(), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Can't create file %s: %s", C.Path.Config(), err.Error())
		}
		// 直接写端口号
		f.Write([]byte(`port: 7890`))
		f.Close()
	}

	// initial mmdb
	if err := initMMDB(); err != nil {
		return fmt.Errorf("Can't initial MMDB: %w", err)
	}
	return nil
}
