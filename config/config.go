package config

import (
	"os"
	"sync"
	"time"

	"github.com/darabuchi/enputi/utils"
	"github.com/darabuchi/log"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
)

const defaultConfigName = "nico.yaml"

var (
	lock sync.RWMutex

	c          = viper.New()
	configPath = defaultConfigName

	changed = atomic.NewBool(false)
)

func init() {
	SetConfigPath(configPath)
	go func(sign chan os.Signal) {
		syncTicker := time.NewTicker(time.Minute)
		defer syncTicker.Stop()

		for {
			select {
			case <-syncTicker.C:
				if changed.Load() {
					Sync()
				}
			case <-sign:
				Sync()
				return
			}
		}
	}(utils.GetExitSign())
}

func Get(key string) any {
	lock.RLock()
	defer lock.RUnlock()

	return c.Get(key)
}

func Set(key string, value any) {
	lock.Lock()
	defer lock.Unlock()

	c.Set(key, value)
	changed.Store(true)
}

func SetConfigPath(filePath string) {
	lock.Lock()
	defer lock.Unlock()

	configPath = filePath

	c.SetConfigFile(configPath)

	err := c.ReadInConfig()
	if err != nil {
		switch e := err.(type) {
		case viper.ConfigFileNotFoundError:
			err = c.WriteConfigAs(configPath)
			if err != nil {
				log.Errorf("err:%v", err)
			}
			log.Debug("not found conf file, use default")
		case *os.PathError:
			err = c.WriteConfigAs(configPath)
			if err != nil {
				log.Errorf("err:%v", err)
			}
			log.Debugf("not find conf file in %s", e.Path)
		default:
			log.Debugf("load config fail:%v", err)
		}
	}
}

func Sync() {
	lock.Lock()
	defer lock.Unlock()

	err := c.WriteConfigAs(configPath)
	if err != nil {
		log.Errorf("err:%v", err)
	} else {
		changed.Store(false)
	}
}
