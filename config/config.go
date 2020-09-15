package config

import (
	"flag"
	"github.com/gisvr/golib/log"
	"io/ioutil"
	"path/filepath"
	"sync/atomic"

	"gopkg.in/yaml.v2"
)

// ConfigFile ..
var (
	Ref          string //指定当前的环境(dev/test/pre)
	VarsFile     string
	ConfigFile   string
	TplFile      string
	ExtFile      string //外部指定变量，yaml语法，用于k8s pod通过docker注入
	globalConfig atomic.Value
)

// Get 返回config实例
func Get() interface{} {
	r := globalConfig.Load()
	if r == nil {
		return nil
	}
	return r
}

// Initialize 会试图从文件中读取config，注意，如果指定了-c ConfigFile，那么defaultConfig中的值会被忽略
func Initialize(config interface{}, defaultConfig interface{}) interface{} {
	if !flag.Parsed() {
		flag.Parse()
	}
	if ConfigFile != "" {
		ConfigFile, _ = filepath.Abs(ConfigFile)
	}
	var dat []byte
	var compile *CompileT
	var err error
	if Ref != "" && VarsFile != "" && TplFile != "" {
		VarsFile, _ = filepath.Abs(VarsFile)
		compile = NewCompileT(Ref)
		err := LoadVars(VarsFile, compile)
		if err != nil {
			log.Fatalf("load vars failed!! err=%s", err.Error())
		}
		dat, err = ioutil.ReadFile(TplFile)
		if err != nil {
			log.Fatalf("load tpl file failed!! err=%s", err.Error())
		}
		cdat, err := ExecTemplateString(string(dat), compile)
		if err != nil {
			log.Errorf("template file failed,err=%s", err)
		} else {
			dat = []byte(cdat)
		}
		if ConfigFile != "" {
			err = ioutil.WriteFile(ConfigFile, dat, 755)
			if err != nil {
				log.Errorf("template file failed,err=%s", err)
			}
		}
	} else if ConfigFile != "" {
		dat, err = ioutil.ReadFile(ConfigFile)
		if err != nil {
			log.Fatalf("load conf file[%s] failed!! err=%s", ConfigFile, err.Error())
		}
	}
	if ExtFile != "" {
		buf, err := ioutil.ReadFile(ExtFile)
		if err != nil {
			log.Fatalf("load conf file[%s] failed!! err=%s", ExtFile, err.Error())
		}
		dat = append(dat, buf...)
	}

	if defaultConfig != nil {
		globalConfig.Store(defaultConfig)
	}
	if dat != nil {
		err = yaml.Unmarshal(dat, config)
		if err != nil {
			log.Fatalf("parse config file %v error. err = %v", ConfigFile, err)
		}
		if err = filterEncrypted(config); err != nil {
			log.Fatalf("error load encrypted config: %v", err)
		}
		globalConfig.Store(config)
	} else {
		log.Infof("config file is null. no -c ??")
	}
	return globalConfig.Load()
}

// Init 传入带有默认值的 config, 加载配置到 config 中
func Init(config interface{}) interface{} {
	return Initialize(config, config)
}
