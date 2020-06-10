package constant

import (
	G "github.com/Dreamacro/clash/log"
	"os"
	P "path"
	"path/filepath"
)

const Name = "clash"

// Path is used to get the configuration path
var Path *path

type path struct {
	homeDir    string
	configFile string
}

// 加载这个文件会初始化
func init() {
	// 获取用户的主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 获取当前目录
		homeDir, _ = os.Getwd()
	}

	// 在主目录下使用.config文件夹
	homeDir = P.Join(homeDir, ".config", Name)
	G.Infoln("homeDir...%s", homeDir)
	// 默认
	Path = &path{homeDir: homeDir, configFile: "config.yaml"}
}

// SetHomeDir is used to set the configuration path
func SetHomeDir(root string) {
	Path.homeDir = root
}

// SetConfig is used to set the configuration file
func SetConfig(file string) {
	Path.configFile = file
}

func (p *path) HomeDir() string {
	return p.homeDir
}

func (p *path) Config() string {
	return p.configFile
}

// Resolve return a absolute path or a relative path with homedir
func (p *path) Resolve(path string) string {
	if !filepath.IsAbs(path) {
		return filepath.Join(p.HomeDir(), path)
	}

	return path
}

func (p *path) MMDB() string {
	return P.Join(p.homeDir, "Country.mmdb")
}
